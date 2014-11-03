package main

import "fmt"
import "math"
import "image"
import "image/png"
import "image/color"
import "os"

type Ray struct {
	origin    *vec3
	direction *vec3
}

type Camera struct {
	location  *vec3
	points_at *vec3
}

type Scene struct {
	primitives []Primitive
	cam        Camera
}

func getColor(ray *Ray, scene *Scene) (r uint8, g uint8, b uint8) {
	var foundPrim Primitive = nil
	color := &vec3{0, 0, 0}
	dist := math.MaxFloat64
	for _, prim := range scene.primitives {
		res, ndist := prim.intersect(ray)
		if res != 0 && ndist < dist {
			foundPrim = prim
			dist = ndist
		}
	}

	if foundPrim == nil {
		return 0, 0, 0
	}

	switch foundPrim.(type) {
	case Light:
		return 255, 255, 255
	default:
		// Intersection point
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
					diff := dot * foundPrim.getDiffuse()
					color = color.add(lPrim.getColor().mult(foundPrim.getColor())).scale(diff)
				}
			}
		}
	}

	return color.toRGB()
}

type WorkUnit struct {
	x, y  float64
	ray   *Ray
	scene *Scene
}
type ResultUnit struct {
	x, y    float64
	r, g, b uint8
}
type Matrix4x4 struct {
	matx [4][4]float64
}

func (m Matrix4x4) multVecMatrix(v *vec3) *vec3 {
	x := v.x*m.matx[0][0] + v.y*m.matx[1][0] + v.z*m.matx[2][0] + m.matx[3][0]
	y := v.x*m.matx[0][1] + v.y*m.matx[1][1] + v.z*m.matx[2][1] + m.matx[3][1]
	z := v.x*m.matx[0][2] + v.y*m.matx[1][2] + v.z*m.matx[2][2] + m.matx[3][2]
	w := v.x*m.matx[0][3] + v.y*m.matx[1][3] + v.z*m.matx[2][3] + m.matx[3][3]
	return &vec3{x / w, y / w, z / w}
}

func (m Matrix4x4) multDirMatrix(v *vec3) *vec3 {
	x := v.x*m.matx[0][0] + v.y*m.matx[1][0] + v.z*m.matx[2][0]
	y := v.x*m.matx[0][1] + v.y*m.matx[1][1] + v.z*m.matx[2][1]
	z := v.x*m.matx[0][2] + v.y*m.matx[1][2] + v.z*m.matx[2][2]
	return &vec3{x, y, z}
}

func main() {
	fmt.Println("HI!")
	outi := image.NewNRGBA(image.Rect(0, 0, 512, 512))
	scene := Scene{
		primitives: []Primitive{Sphere{1, &vec3{0, 0, -1}, &Material{&vec3{1, 0, 0}, 0.5, 0.8}},
			Sphere{0.3, &vec3{0, 0, 0}, &Material{&vec3{0, 1, 0}, 0, 1}},
			Plane{&vec3{2, 0, 0}, &vec3{-1, 0, 0}, &Material{&vec3{0, 0, 1}, 0, 1}},
			Plane{&vec3{-2, 0, 0}, &vec3{1, 0, 0}, &Material{&vec3{1, 0, 1}, 0, 1}},
			Plane{&vec3{0, 0, -2}, &vec3{0, 0, 1}, &Material{&vec3{1, 1, 1}, 0, 1}},
			Plane{&vec3{0, -2, 0}, &vec3{0, 1, 0}, &Material{&vec3{1, 1, 1}, 0, 1}},
			Light{&Sphere{0.1, &vec3{0, 0.5, 2}, &Material{&vec3{1, 1, 1}, 0, 0}}, &Material{&vec3{1, 1, 1}, 0, 0}}},
		cam: Camera{&vec3{0, 0, 3}, &vec3{0, 0, 0}},
	}

	imgSize := outi.Bounds().Size().X * outi.Bounds().Size().Y
	jobs := make(chan WorkUnit, imgSize)
	results := make(chan ResultUnit, imgSize)
	screenX := float64(outi.Bounds().Size().X)
	screenY := float64(outi.Bounds().Size().Y)
	aspectRatio := screenX / screenY
	fovDeg := 60.0
	// Convert to radians and divide by two
	angle := math.Tan(fovDeg * math.Pi / 180 / 2)
	var c2wmatrix [4][4]float64
	c2wmatrix[0][0] = 1
	c2wmatrix[1][1] = 1
	c2wmatrix[2][2] = 1
	c2wmatrix[3][3] = 1
	cam2World := Matrix4x4{c2wmatrix}

	for w := 1; w <= 8; w++ {
		go func() {
			for wu := range jobs {
				r, g, b := getColor(wu.ray, wu.scene)
				results <- ResultUnit{wu.x, wu.y, r, g, b}
			}
		}()
	}

	origin := scene.cam.location.get_copy()
	origin = cam2World.multVecMatrix(origin)
	for x := 0.0; x < screenX; x++ {
		for y := 0.0; y < screenY; y++ {
			// Get pixel in terms of world space on the scene plane
			u := (2*((x+0.5)/screenX) - 1) * angle * aspectRatio
			v := (1 - 2*((y+0.5)/screenY)) * angle
			rayD := cam2World.multDirMatrix(&vec3{u, v, -1}).normalize()
			ray := Ray{origin, rayD}
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
