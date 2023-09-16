package internal

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
	length := 2 * Dot(hi.point, hi.normal)
	n := hi.normal.Cpy()
	n.Scale(length)
	reflected := Sub(hi.point, n)

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
