package main

import "math"

type Primitive interface {
	// 1 = hit, 0 = miss, -1 = ray began inside primitive
	intersect(ray *Ray) (int, float64)
	getCenter() *vec3
	getNormal(p *vec3) *vec3
	MatI
}

type MatI interface {
	getColor() *vec3
	getDiffuse() float64
	getSpecular() float64
	getReflectivity() float64
}

type Material struct {
	color    *vec3
	reflect  float64
	diffuse  float64
	specular float64
}

func (m Material) getColor() *vec3 {
	return m.color
}
func (m Material) getDiffuse() float64 {
	return m.diffuse
}
func (m Material) getSpecular() float64 {
	return m.specular
}
func (m Material) getReflectivity() float64 {
	return m.reflect
}

type Sphere struct {
	radius float64
	center *vec3
	*Material
}

func (s Sphere) intersect(ray *Ray) (int, float64) {
	v := ray.origin.subtract(s.center)
	b := -v.dot(ray.direction)
	dist := math.MaxFloat64
	if b < 0 {
		return 0, dist
	}
	det := (b * b) - v.dot(v) + (s.radius * s.radius)
	retval := 0
	if det > 0 {
		det = math.Sqrt(det)
		i1 := b - det
		i2 := b + det
		if i2 > 0 {
			if i1 < 0 {
				retval = -1
				dist = i2
			} else {
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

func (s Sphere) getNormal(p *vec3) *vec3 {
	return p.subtract(s.center).scale(1.0 / s.radius)
}

type Plane struct {
	origin *vec3
	normal *vec3
	*Material
}

func (p Plane) intersect(ray *Ray) (int, float64) {
	hit := 0
	dist := 0.0
	denom := p.normal.dot(ray.direction)
	if denom != 0 {
		dist = p.normal.dot(p.origin.subtract(ray.origin)) / denom
		if dist > 0 {
			hit = 1
		}
	}
	return hit, dist
}
func (p Plane) getCenter() *vec3 {
	return p.origin
}
func (p Plane) getNormal(p1 *vec3) *vec3 {
	return p.normal
}

type Light struct {
	emitter Primitive
	*Material
}

func (l Light) intersect(ray *Ray) (int, float64) {
	return l.emitter.intersect(ray)
}
func (l Light) getCenter() *vec3 {
	return l.emitter.getCenter()
}
func (l Light) getNormal(p *vec3) *vec3 {
	return l.emitter.getNormal(p).normalize()
}
