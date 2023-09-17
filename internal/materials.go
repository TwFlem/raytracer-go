package internal

import "math"

type Material interface {
	Scatter(r *Ray, hi HitInfo) (ScatterInfo, bool)
}

type ScatterInfo struct {
	ray         Ray
	attenuation Vec3[float32]
}

type Lambertian struct {
	albedo Vec3[float32]
}

func NewLambertian(albedo Vec3[float32]) Lambertian {
	return Lambertian{
		albedo: albedo,
	}
}

func (l *Lambertian) Scatter(r *Ray, hi HitInfo) (ScatterInfo, bool) {
	dir := Add(hi.normal, NewVec3UnitRandOnUnitSphere32())
	if dir.NearZero() {
		dir = hi.normal
	}
	return ScatterInfo{
		ray:         *NewRay(hi.point, dir),
		attenuation: l.albedo,
	}, true
}

type Metal struct {
	albedo Vec3[float32]
	fuzz   float32
}

func NewMetal(albedo Vec3[float32], fuzz float32) Metal {
	return Metal{
		albedo: albedo,
		fuzz:   fuzz,
	}
}

func (m *Metal) Scatter(r *Ray, hi HitInfo) (ScatterInfo, bool) {
	reflected := reflect(hi.point, hi.normal)

	fuzz := NewVec3UnitRandOnUnitSphere32()
	fuzz.Scale(m.fuzz)

	scattered := Add(reflected, fuzz)
	if Dot(scattered, hi.normal) > 0 {
		return ScatterInfo{
			ray:         *NewRay(hi.point, scattered),
			attenuation: m.albedo,
		}, true
	}
	return ScatterInfo{}, false
}

type Refractor interface {
	GetRefractionIndex() float32
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
	etaOEtaPrime := float32(1.0 / d.refractiveIndex)
	if !hi.frontFace {
		etaOEtaPrime = d.refractiveIndex
	}
	if refractor, ok := hi.material.(Refractor); ok {
		etaOEtaPrime *= refractor.GetRefractionIndex()
	}

	unitDir := r.dir.Cpy()
	unitDir.Unit()

	cosTheta := float32(math.Min(float64(Dot(Scale(unitDir, -1), hi.normal)), 1.0))
	sinTheta := float32(math.Sqrt(1 - float64(cosTheta*cosTheta)))
	if sinTheta*etaOEtaPrime > 1.0 {
		reflected := reflect(hi.point, hi.normal)
		return ScatterInfo{
			ray:         *NewRay(hi.point, reflected),
			attenuation: NewVec3[float32](1, 1, 1),
		}, true
	}
	refracted := refract(unitDir, hi.normal, etaOEtaPrime)
	return ScatterInfo{
		ray:         *NewRay(hi.point, refracted),
		attenuation: NewVec3[float32](1, 1, 1),
	}, true

}
