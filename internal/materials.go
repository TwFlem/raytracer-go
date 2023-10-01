package internal

import (
	"math"
	"math/rand"
)

type Material interface {
	Scatter(r *Ray, hi HitInfo) (ScatterInfo, bool)
}

type ScatterInfo struct {
	ray         Ray
	attenuation Vec3
}

type Lambertian struct {
	albedo Vec3
}

func NewLambertian(albedo Vec3) Lambertian {
	return Lambertian{
		albedo: albedo,
	}
}

func (l *Lambertian) Scatter(r *Ray, hi HitInfo) (ScatterInfo, bool) {
	dir := Add(hi.normal, NewVec3UnitRandOnUnitSphere32(r.rand))
	if dir.NearZero() {
		dir = hi.normal
	}
	return ScatterInfo{
		ray:         *NewRay(hi.point, dir, r.rand),
		attenuation: l.albedo,
	}, true
}

type Metal struct {
	albedo Vec3
	fuzz   float32
}

func NewMetal(albedo Vec3, fuzz float32) Metal {
	return Metal{
		albedo: albedo,
		fuzz:   fuzz,
	}
}

func (m *Metal) Scatter(r *Ray, hi HitInfo) (ScatterInfo, bool) {
	unitDir := Unit(r.dir)
	reflected := reflect(unitDir, hi.normal)

	fuzz := NewVec3UnitRandOnUnitSphere32(r.rand)
	fuzz.Scale(m.fuzz)

	scattered := Add(reflected, fuzz)
	if Dot(scattered, hi.normal) > 0 {
		return ScatterInfo{
			ray:         *NewRay(hi.point, scattered, r.rand),
			attenuation: m.albedo,
		}, true
	}
	return ScatterInfo{}, false
}

type Dielectric struct {
	refractiveIndex float32
}

func NewDielectric(refracitveIndex float32) Dielectric {
	return Dielectric{
		refractiveIndex: refracitveIndex,
	}
}

func (d *Dielectric) Scatter(r *Ray, hi HitInfo) (ScatterInfo, bool) {
	etaOEtaPrime := d.refractiveIndex
	if hi.frontFace {
		etaOEtaPrime = 1.0 / d.refractiveIndex
	}

	unitDir := Unit(r.dir)

	cosTheta := float32(math.Min(float64(Dot(Scale(unitDir, -1), hi.normal)), 1.0))
	sinTheta := float32(math.Sqrt(1 - float64(cosTheta*cosTheta)))
	cannotRefract := sinTheta*etaOEtaPrime > 1.0
	var direction Vec3
	if cannotRefract || reflectance(cosTheta, etaOEtaPrime) > rand.Float32() {
		direction = reflect(unitDir, hi.normal)
	} else {
		direction = refract(unitDir, hi.normal, etaOEtaPrime)
	}

	return ScatterInfo{
		ray:         *NewRay(hi.point, direction, r.rand),
		attenuation: NewVec3(1, 1, 1, 0),
	}, true
}

func reflectance(cosTheta, etaOEtaPrime float32) float32 {
	r0 := (1.0 - etaOEtaPrime) / (1.0 + etaOEtaPrime)
	r0 *= r0
	return r0 + (1-r0)*float32(math.Pow(1-float64(cosTheta), 5))
}
