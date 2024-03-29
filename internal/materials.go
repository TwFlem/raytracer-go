package internal

import (
	"image"
	"math"
	"math/rand"
)

type Material interface {
	Scatter(r *Ray, hi HitInfo) (ScatterInfo, bool)
	Emit(u, v float32, p Vec3) Color
}

type ScatterInfo struct {
	ray         Ray
	attenuation Color
}

type Lambertian struct {
	albedo Texture
}

func (l Lambertian) Emit(u float32, v float32, p Vec3) Color {
	return NewVec3Zero()
}

func NewLambertian(albedo Texture) Lambertian {
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
		attenuation: l.albedo.GetTexture(hi.u, hi.v, hi.point),
	}, true
}

type Metal struct {
	albedo Color
	fuzz   float32
}

func (m Metal) Emit(u float32, v float32, p Vec3) Color {
	return NewVec3Zero()
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

func (d *Dielectric) Emit(u float32, v float32, p Vec3) Color {
	return NewVec3Zero()
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
		attenuation: NewVec3(1, 1, 1),
	}, true
}

func reflectance(cosTheta, etaOEtaPrime float32) float32 {
	r0 := (1.0 - etaOEtaPrime) / (1.0 + etaOEtaPrime)
	r0 *= r0
	return r0 + (1-r0)*float32(math.Pow(1-float64(cosTheta), 5))
}

type Checkered struct {
	scale float32
	even  Color
	odd   Color
}

func (c *Checkered) GetTexture(u float32, v float32, point Vec3) Color {
	invScale := 1 / c.scale
	x := int(math.Floor(float64(invScale * point.X)))
	y := int(math.Floor(float64(invScale * point.Y)))
	z := int(math.Floor(float64(invScale * point.Z)))

	if (x+y+z)%2 == 0 {
		return c.even
	}
	return c.odd
}

func NewCheckered(scale float32, even, odd Vec3) Checkered {
	return Checkered{
		scale: scale,
		even:  even,
		odd:   odd,
	}
}

type Texture interface {
	GetTexture(u, v float32, point Vec3) Color
}

type SolidColor struct {
	albedo Color
}

func (s SolidColor) GetTexture(u float32, v float32, point Vec3) Color {
	return s.albedo.GetColor()
}

func NewSolidColor(x, y, z float32) SolidColor {
	return SolidColor{
		albedo: NewVec3(x, y, z),
	}
}

type ImageTexture struct {
	img image.Image
}

func NewImageTexture(img image.Image) ImageTexture {
	return ImageTexture{
		img: img,
	}
}

func (it *ImageTexture) GetTexture(u float32, v float32, point Vec3) Color {
	// Debug color if there is no height to the img
	if it.img.Bounds().Dy() <= 0 {
		return NewVec3(0, 1, 1)
	}

	u = Clamp(0, 1, u)
	v = 1 - Clamp(0, 1, v)

	i := u * float32(it.img.Bounds().Dx())
	j := v * float32(it.img.Bounds().Dy())
	pixel := it.img.At(int(i), int(j))

	colScale := float32(1.0 / 65535.0)
	r, g, b, _ := pixel.RGBA()

	vec := NewVec3(float32(r)*colScale, float32(g)*colScale, float32(b)*colScale)
	return vec
}

type Perlin struct {
	randVec3 []Vec3
	permX    []int
	permY    []int
	permZ    []int
}

func NewPerlin(randCtx *rand.Rand) Perlin {
	pointCount := 256
	points := make([]Vec3, pointCount)
	for i := 0; i < len(points); i++ {
		points[i] = NewVec3RandRange32(randCtx, -1, 1)
	}

	return Perlin{
		randVec3: points,
		permX:    Permute(GetNums(pointCount)),
		permY:    Permute(GetNums(pointCount)),
		permZ:    Permute(GetNums(pointCount)),
	}

}

func smoothstep(t float32) float32 {
	return t * t * (3 - 2*t)
}

// TODO: Study noise
func (per *Perlin) Noise(p Vec3) float32 {
	xi := float32(math.Floor(float64(p.X)))
	yi := float32(math.Floor(float64(p.Y)))
	zi := float32(math.Floor(float64(p.Z)))

	tx := p.X - float32(xi)
	ty := p.Y - float32(yi)
	tz := p.Z - float32(zi)

	rx0 := int(xi) & 255
	rx1 := (rx0 + 1) & 255
	ry0 := int(yi) & 255
	ry1 := (ry0 + 1) & 255
	rz0 := int(zi) & 255
	rz1 := (rz0 + 1) & 255

	c000 := Dot(per.randVec3[per.permX[rx0]^per.permY[ry0]^per.permZ[rz0]], NewVec3(tx, ty, tz))
	c001 := Dot(per.randVec3[per.permX[rx0]^per.permY[ry0]^per.permZ[rz1]], NewVec3(tx, ty, tz-1))
	c010 := Dot(per.randVec3[per.permX[rx0]^per.permY[ry1]^per.permZ[rz0]], NewVec3(tx, ty-1, tz))
	c011 := Dot(per.randVec3[per.permX[rx0]^per.permY[ry1]^per.permZ[rz1]], NewVec3(tx, ty-1, tz-1))
	c100 := Dot(per.randVec3[per.permX[rx1]^per.permY[ry0]^per.permZ[rz0]], NewVec3(tx-1, ty, tz))
	c101 := Dot(per.randVec3[per.permX[rx1]^per.permY[ry0]^per.permZ[rz1]], NewVec3(tx-1, ty, tz-1))
	c110 := Dot(per.randVec3[per.permX[rx1]^per.permY[ry1]^per.permZ[rz0]], NewVec3(tx-1, ty-1, tz))
	c111 := Dot(per.randVec3[per.permX[rx1]^per.permY[ry1]^per.permZ[rz1]], NewVec3(tx-1, ty-1, tz-1))

	return TriLinearLerp(smoothstep(tx), smoothstep(ty), smoothstep(tz), c000, c100, c010, c110, c001, c101, c011, c111)
}

func (per *Perlin) Turb(p Vec3, depth int) float32 {
	sum := float32(0)
	weight := float32(1.0)

	for i := 0; i < depth; i++ {
		sum += weight * per.Noise(p)
		weight *= 0.5
		p.Scale(2)
	}

	return float32(math.Abs(float64(sum)))
}

func GetNums(closedUpperEnd int) []int {
	n := make([]int, closedUpperEnd)
	for i := 0; i < len(n); i++ {
		n[i] = i
	}
	return n
}

func Permute(p []int) []int {
	for i := len(p) - 1; i > 0; i-- {
		target := rand.Intn(i)
		p[i], p[target] = p[target], p[i]
	}
	return p
}

type NoiseTexture struct {
	perlin Perlin
	scale  float32
}

func (n *NoiseTexture) GetTexture(u float32, v float32, point Vec3) Color {
	point.Scale(n.scale)
	return Scale(NewVec3Unit(), 0.5*(1+float32(math.Sin(float64(point.Z+10*n.perlin.Turb(point, 7))))))
}

func NewNoiseTexture(randCtx *rand.Rand, scale float32) NoiseTexture {
	return NoiseTexture{
		perlin: NewPerlin(randCtx),
		scale:  scale,
	}
}

type DiffuseLight struct {
	emit Texture
}

func (d DiffuseLight) Scatter(r *Ray, hi HitInfo) (ScatterInfo, bool) {
	return ScatterInfo{}, false
}

func NewDiffuseLight(emit Texture) DiffuseLight {
	return DiffuseLight{
		emit: emit,
	}
}

func (d DiffuseLight) Emit(u float32, v float32, p Vec3) Color {
	return d.emit.GetTexture(u, v, p)
}
