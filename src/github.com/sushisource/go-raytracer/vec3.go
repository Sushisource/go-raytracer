package main

import "math"

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
