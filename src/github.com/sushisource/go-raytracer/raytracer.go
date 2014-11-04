package main

import "fmt"
import "math"
import "image"
import "image/png"
import "image/color"
import "os"

type Ray struct {
	origin      *vec3
	direction   *vec3
	maxDistance float64
}

type Camera struct {
	location  *vec3
	points_at *vec3
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

type Scene struct {
	primitives []Primitive
	cam        Camera
}

func raytrace(ray *Ray, scene *Scene, curDepth int) *vec3 {
	var foundPrim Primitive = nil
	color := &vec3{0, 0, 0}
	ray.maxDistance = math.MaxFloat64
	for _, prim := range scene.primitives {
		res, _ := prim.intersect(ray)
		if res != 0 {
			foundPrim = prim
		}
	}

	if foundPrim == nil {
		return &vec3{0, 0, 0}
	}

	switch foundPrim.(type) {
	case Light:
		return foundPrim.getColor()
	default:
		// Intersection point
		pi := ray.origin.add(ray.direction.scale(ray.maxDistance))
		n := foundPrim.getNormal(pi)
		//TODO: Would be better to maintain a separate list of lights
		for _, lPrim := range scene.primitives {
			switch lPrim.(type) {
			case Light:
				shade := 1.0
				// Get direction to light
				l := lPrim.getCenter().subtract(pi)
				tdist := l.length()
				l.normalize()
				// Shadow TODO: Will need to expand for different light types
				shadeRay := &Ray{pi.add(l.scale(0.0001)), l, tdist}
				for _, sPrim := range scene.primitives {
					switch sPrim.(type) {
					case Light:
						continue
					default:
						res, _ := sPrim.intersect(shadeRay)
						if res != 0 {
							shade = 0
							break
						}
					}
				}
				// Diffuse
				dot := l.dot(n)
				if dot > 0 {
					diff := dot * foundPrim.getDiffuse() * shade
					color = color.add(lPrim.getColor().mult(foundPrim.getColor()).scale(diff))
				}
				// Specular
				Vs := ray.direction
				Rs := l.subtract(n.scale(2 * dot))
				dot = Vs.dot(Rs)
				if dot > 0 {
					spec := math.Pow(dot, 20) * foundPrim.getSpecular() * shade
					color = color.add(lPrim.getColor().scale(spec))
				}
			}
		}
		// Reflection
		reflectivity := foundPrim.getReflectivity()
		if reflectivity > 0 {
			R := ray.direction.subtract(n.scale(2 * ray.direction.dot(n)))
			if curDepth < 8 { // TODO: Put trace depth and eps values in constants
				// epsilon val is small
				rRay := &Ray{pi.add(R.scale(0.0001)), R, 0}
				rCol := raytrace(rRay, scene, curDepth+1)
				color = color.add(rCol.mult(foundPrim.getColor()).scale(reflectivity))
			}
		}
	}

	return color
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

func main() {
	outi := image.NewNRGBA(image.Rect(0, 0, 512, 512))
	scene := Scene{
		primitives: []Primitive{
			Sphere{0.3, &vec3{0, 0, 0}, &Material{&vec3{0, 1, 0}, 0.1, 0.8, 0.5}},
			Sphere{0.1, &vec3{-.8, 0, 0}, &Material{&vec3{0, 1, 0}, 0.1, 0.9, 0.5}},
			Sphere{0.3, &vec3{1, -1.1, -.41}, &Material{&vec3{0, 1, 0}, 0.1, 0.8, 0.5}},
			Sphere{.6, &vec3{-1, 1, -1}, &Material{&vec3{1, 1, 1}, 0.8, 0.4, 0.5}},
			Sphere{.6, &vec3{-1, -1, -1}, &Material{&vec3{1, 1, 1}, 0.8, 0.4, 0.5}},
			Plane{&vec3{2, 0, 0}, &vec3{-1, 0, 0}, &Material{&vec3{0, 0, 1}, 0, 1, 0}},
			Plane{&vec3{-2, 0, 0}, &vec3{1, 0, 0}, &Material{&vec3{1, 0, 1}, 0, 1, 0}},
			Plane{&vec3{0, 0, -2}, &vec3{0, 0, 1}, &Material{&vec3{1, 1, 1}, 0, 1, 0}},
			Plane{&vec3{0, -2, 0}, &vec3{0, 1, 0}, &Material{&vec3{1, 1, 1}, 0, 1, 0}},
			Light{&Sphere{0.05, &vec3{1.0, 0, 0}, &Material{&vec3{1, 1, 1}, 0, 0, 0}}, &Material{&vec3{1, 1, 1}, 0, 0, 0}},
			// Light{&Sphere{0.1, &vec3{1.0, 1.0, 0}, &Material{&vec3{1, 1, 1}, 0, 0, 0}}, &Material{&vec3{1, 0, 1}, 0, 0, 0}},
			// Light{&Sphere{0.1, &vec3{1.0, -1.0, 0}, &Material{&vec3{1, 1, 1}, 0, 0, 0}}, &Material{&vec3{0, 1, 1}, 0, 0, 0}},
		},
		cam: Camera{&vec3{0, 0, 3}, &vec3{0, 0, 0}},
	}

	fmt.Println("Starting!")

	imgSize := outi.Bounds().Size().X * outi.Bounds().Size().Y
	jobs := make(chan WorkUnit, imgSize)
	results := make(chan ResultUnit, imgSize)
	screenX := float64(outi.Bounds().Size().X)
	screenY := float64(outi.Bounds().Size().Y)
	aspectRatio := screenX / screenY
	fovDeg := 70.0
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
				r, g, b := raytrace(wu.ray, wu.scene, 0).toRGB()
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
			ray := Ray{origin, rayD, 0}
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
	fmt.Println("done!")
}
