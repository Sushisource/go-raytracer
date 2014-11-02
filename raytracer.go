package main

import "fmt"
import "math"
import "image"
import "image/png"
import "image/color"
import "os"

type vec3 struct {
	x, y, z float64
}

func (v *vec3) length() float64 {
	return math.Sqrt(v.x*v.x + v.y*v.y + v.z*v.z)
}

// MUTATING METHODS -----------------------------------------------------------
func (v *vec3) normalize() *vec3 {
	l := 1.0 / v.length()
	v.x *= l
	v.y *= l
	v.z *= l
	return v
}

func (v *vec3) neg() *vec3 {
	v.x = -v.x
	v.y = -v.y
	v.z = -v.z
	return v
}

// COPYING METHODS ------------------------------------------------------------
func (v *vec3) get_copy() *vec3 {
	return &vec3{v.x, v.y, v.z}
}

func (v *vec3) subtract(v2 *vec3) *vec3 {
	return &vec3{v.x - v2.x, v.y - v2.y, v.z - v2.z}
}

func (v *vec3) add(v2 *vec3) *vec3 {
	return &vec3{v.x + v2.x, v.y + v2.y, v.z + v2.z}
}

func (v *vec3) dot(v2 *vec3) float64 {
	return v.x*v2.x + v.y*v2.y + v.z*v2.z
}

func (v *vec3) scale(mag float64) *vec3 {
	return &vec3{v.x * mag, v.y * mag, v.z * mag}
}

func (v *vec3) mult(v2 *vec3) *vec3 {
	return &vec3{v.x * v2.x, v.y * v2.y, v.z * v2.z}
}

func (v *vec3) toRGB() (uint8, uint8, uint8) {
	return uint8(v.x * 255), uint8(v.y * 255), uint8(v.z * 255)
}

type Ray struct {
	origin    *vec3
	direction *vec3
}

type Camera struct {
	location  *vec3
	points_at *vec3
}

type Primitive interface {
	// 1 = hit, 0 = miss, -1 = ray began inside primitive
	intersect(ray *Ray) (int, float64)
	getCenter() *vec3
	getColor() *vec3
	getNormal(p *vec3) *vec3
}

type Scene struct {
	primitives []Primitive
	cam        Camera
}

type Sphere struct {
	radius float64
	center *vec3
	color  *vec3
}

func (s Sphere) intersect(ray *Ray) (int, float64) {
	v := ray.origin.subtract(s.center)
	b := -v.dot(ray.direction)
	dist := math.MaxFloat64
	det := (b * b) - v.dot(v) + (s.radius * s.radius)
	retval := 0
	if det > 0 {
		det = math.Sqrt(det)
		i1 := b - det
		i2 := b + det
		if i2 > 0 {
			if i1 < 0 {
				if i2 < dist {
					retval = -1
					dist = i2
				}
			} else if i1 < dist {
				retval = 1
				dist = i1
			}
		}
	}
	return retval, dist
}

func (s Sphere) getCenter() *vec3 {
	return s.center
}

func (s Sphere) getColor() *vec3 {
	return s.color
}

func (s Sphere) getNormal(p *vec3) *vec3 {
	return p.subtract(s.center).scale(1.0 / s.radius)
}

type Plane struct {
	origin *vec3
	normal *vec3
}

type Light struct {
	emitter Primitive
	color   *vec3
}

func (l Light) intersect(ray *Ray) (int, float64) {
	return l.emitter.intersect(ray)
}
func (l Light) getColor() *vec3 {
	return l.color
}
func (l Light) getCenter() *vec3 {
	return l.emitter.getCenter()
}
func (l Light) getNormal(p *vec3) *vec3 {
	return l.emitter.getNormal(p)
}

func getColor(ray *Ray, scene *Scene) (r uint8, g uint8, b uint8) {
	var foundPrim Primitive = nil
	color := &vec3{0, 0, 0}
	dist := math.MaxFloat64
	res := 0
	for _, prim := range scene.primitives {
		res, dist = prim.intersect(ray)
		if res != 0 {
			foundPrim = prim
			break //TODO: Implement actual depth
		}
	}

	if foundPrim == nil {
		return 0, 0, 0
	}

	switch foundPrim.(type) {
	case Light:
		return 255, 255, 255
	default:
		pi := ray.origin.add(ray.direction.scale(dist))
		//TODO: Would be better to maintain a separate list of lights
		for _, lPrim := range scene.primitives {
			switch lPrim.(type) {
			case Light:
				// Get direction to light
				l := lPrim.getCenter().subtract(pi)
				l.normalize()
				n := foundPrim.getNormal(pi)
				dot := n.dot(l)
				if dot > 0 {
					// TODO: Use real materials
					diff := dot * 0.8
					color = color.add(lPrim.getColor().mult(foundPrim.getColor())).scale(diff)
				}
			}
		}
	}

	return color.toRGB()
}

type WorkUnit struct {
	x, y  int32
	ray   *Ray
	scene *Scene
}
type ResultUnit struct {
	x, y    int32
	r, g, b uint8
}

func main() {
	fmt.Println("HI!")
	outi := image.NewNRGBA(image.Rect(0, 0, 512, 512))
	scene := Scene{
		primitives: []Primitive{Sphere{1, &vec3{-2, 0, 20}, &vec3{1, 0, 0}},
			Sphere{1, &vec3{-2.5, 0.5, 5}, &vec3{0, 1, 0}},
			Light{&Sphere{0.3, &vec3{-1, -1, 4}, &vec3{1, 1, 1}}, &vec3{1, 1, 1}}},
		cam: Camera{&vec3{0, 0, -5}, &vec3{0, 0, 0}},
	}
	//TODO: Use this
	//scenePlane := Plane{&vec3{0,0,0}, &vec3{0,0,1}}

	imgSize := outi.Bounds().Size().X * outi.Bounds().Size().Y
	jobs := make(chan WorkUnit, imgSize)
	results := make(chan ResultUnit, imgSize)

	for w := 1; w <= 8; w++ {
		go func() {
			for wu := range jobs {
				r, g, b := getColor(wu.ray, wu.scene)
				results <- ResultUnit{wu.x, wu.y, r, g, b}
			}
		}()
	}

	for x := int32(0); x < int32(outi.Bounds().Size().X); x++ {
		for y := int32(0); y < int32(outi.Bounds().Size().Y); y++ {
			origin := scene.cam.location.get_copy()
			// Get pixel in terms of world space on the scene plane
			u := float64(x) / float64(outi.Bounds().Size().X)
			v := float64(y) / float64(outi.Bounds().Size().Y)
			ray := Ray{origin, (&vec3{-1.0 * 2.0 * u, -1.0 + 2.0*v, 0}).subtract(origin).normalize()}
			jobs <- WorkUnit{x, y, &ray, &scene}
		}
	}
	close(jobs)

	for a := 1; a <= imgSize; a++ {
		res := <-results
		outi.Set(int(res.x), int(res.y), color.NRGBA{res.r, res.g, res.b, 255})
	}

	toimg, _ := os.Create("output.png")
	defer toimg.Close()

	png.Encode(toimg, outi)
	fmt.Println("done")
}
