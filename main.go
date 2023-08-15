// Copyright 2018 The Ebiten Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"math/rand"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"

	fonts "github.com/ShaolingPu/battleCity/resources/fonts/tank"
	resources "github.com/ShaolingPu/battleCity/resources/images/tank"
	levels "github.com/ShaolingPu/battleCity/resources/levels/tank"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Entity interface {
	GetInfo() (Width, Height int, X, Y float64)
}

var (
	levels_enemies [35][4]int
	born_positions [3][2]int
	max_enemies    int
)

const (
	screenWidth   = 416
	screenHeight  = 416
	tileSize      = 16
	fontSize      = 24
	titleFontSize = fontSize * 1.5
	smallFontSize = fontSize / 2
)

var (
	tilesImage   *ebiten.Image
	brickImage   *ebiten.Image
	steelImage   *ebiten.Image
	grassImage   *ebiten.Image
	waterImage   *ebiten.Image
	player1Image *ebiten.Image
	player2Image *ebiten.Image
	enemy1Image  *ebiten.Image
	enemy2Image  *ebiten.Image
	enemy3Image  *ebiten.Image
	enemy4Image  *ebiten.Image
	enemy5Image  *ebiten.Image
	enemy6Image  *ebiten.Image
	enemy7Image  *ebiten.Image
	enemy8Image  *ebiten.Image
)

var (
	titleArcadeFont font.Face
	arcadeFont      font.Face
	smallArcadeFont font.Face
)

func init() {
	// Decode an image from the image file's byte slice.
	img, _, err := image.Decode(bytes.NewReader(resources.Sprites))
	if err != nil {
		log.Fatal(err)
	}
	tilesImage = ebiten.NewImageFromImage(img)
	levels_enemies = [35][4]int{{18, 2, 0, 0}, {14, 4, 0, 2}, {14, 4, 0, 2}, {2, 5, 10, 3}, {8, 5, 5, 2},
		{9, 2, 7, 2}, {7, 4, 6, 3}, {7, 4, 7, 2}, {6, 4, 7, 3}, {12, 2, 4, 2},
		{5, 5, 4, 6}, {0, 6, 8, 6}, {0, 8, 8, 4}, {0, 4, 10, 6}, {0, 2, 10, 8},
		{16, 2, 0, 2}, {8, 2, 8, 2}, {2, 8, 6, 4}, {4, 4, 4, 8}, {2, 8, 2, 8},
		{6, 2, 8, 4}, {6, 8, 2, 4}, {0, 10, 4, 6}, {10, 4, 4, 2}, {0, 8, 2, 10},
		{4, 6, 4, 6}, {2, 8, 2, 8}, {15, 2, 2, 1}, {0, 4, 10, 6}, {4, 8, 4, 4},
		{3, 8, 3, 6}, {6, 4, 2, 8}, {4, 4, 4, 8}, {0, 10, 4, 6}, {0, 6, 4, 10},
	}
	born_positions = [3][2]int{{3, 3}, {192, 3}, {381, 3}}
	max_enemies = 4
	// brickImage = tilesImage.SubImage(image.Rect(48, 64, tileSize/2, tileSize/2)).(*ebiten.Image)
	// steelImage = tilesImage.SubImage(image.Rect(48, 72, tileSize/2, tileSize/2)).(*ebiten.Image)
	// grassImage = tilesImage.SubImage(image.Rect(56, 72, tileSize/2, tileSize/2)).(*ebiten.Image)
	// waterImage = tilesImage.SubImage(image.Rect(64, 64, tileSize/2, tileSize/2)).(*ebiten.Image)
}

func init() {
	tt, err := opentype.Parse(fonts.PrStart_ttf)
	if err != nil {
		log.Fatal(err)
	}
	const dpi = 72
	titleArcadeFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    titleFontSize,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	arcadeFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	smallArcadeFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    smallFontSize,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
}

type Mode int

const (
	ModeTitle Mode = iota
	ModeGame
	ModeGameOver
)

type Bullet struct {
	Image       *ebiten.Image
	X           float64
	Y           float64
	F           int
	SpeedFactor float64
	Owner       *Tank
}

type Tank struct {
	enemy       bool
	Image       *ebiten.Image
	X           float64
	Y           float64
	Face        int
	SpeedFactor float64
	Failed      bool
}

type Other struct {
	Image *ebiten.Image
	X     float64
	Y     float64
	T     int
}

func (o *Other) GetInfo() (Width, Height int, X, Y float64) {
	w, h := o.Image.Bounds().Dx(), o.Image.Bounds().Dy()
	return w * 2, h * 2, o.X, o.Y
}

func (t *Tank) Turn(i int) {
	t.Face = i
}

func (t *Tank) GetInfo() (Width, Height int, X, Y float64) {
	w, h := t.Image.Bounds().Dx(), t.Image.Bounds().Dy()
	return w * 2, h * 2, t.X, t.Y
}

func (b *Bullet) Move() {
	switch b.F {
	case 0:
		b.Y -= b.SpeedFactor
	case 1:
		b.X += b.SpeedFactor
	case 2:
		b.Y += b.SpeedFactor
	default:
		b.X -= b.SpeedFactor
	}
}

func RectCollision(aWidth, aHeight int, aX, aY float64, bWidth, bHeight int, bX, bY float64, eqa bool) bool {
	top, left := aY, aX
	bottom, right := aY+float64(aHeight), aX+float64(aWidth)
	// top left
	x, y := bX, bY
	if eqa && y >= top && y <= bottom && x >= left && x <= right {
		return true
	} else if y > top && y < bottom && x > left && x < right {
		return true
	}

	// top right
	x, y = bX+float64(bWidth), bY
	if eqa && y >= top && y <= bottom && x >= left && x <= right {
		return true
	} else if y > top && y < bottom && x > left && x < right {
		return true
	}

	// bottom left
	x, y = bX, bY+float64(bHeight)
	if eqa && y >= top && y <= bottom && x >= left && x <= right {
		return true
	} else if y > top && y < bottom && x > left && x < right {
		return true
	}

	// bottom right
	x, y = bX+float64(bWidth), bY+float64(bHeight)
	if eqa && y >= top && y <= bottom && x >= left && x <= right {
		return true
	} else if y > top && y < bottom && x > left && x < right {
		return true
	}

	return false
}

func CheckCollision(A, B Entity, eqa bool) bool {
	aWidth, aHeight, aX, aY := A.GetInfo()
	bWidth, bHeight, bX, bY := B.GetInfo()
	return RectCollision(aWidth, aHeight, aX, aY, bWidth, bHeight, bX, bY, eqa)

}

func (b *Bullet) GetInfo() (Width, Height int, X, Y float64) {
	w, h := b.Image.Bounds().Dx(), b.Image.Bounds().Dy()
	return w * 2, h * 2, b.X, b.Y
}

func (t *Tank) Fire() *Bullet {
	w, h := float64(t.Image.Bounds().Dx()), float64(t.Image.Bounds().Dy())
	img := tilesImage.SubImage(image.Rect(75, 74, 75+3, 74+4)).(*ebiten.Image)
	dx, dy := float64(img.Bounds().Dx()), float64(img.Bounds().Dy())

	var x, y float64
	switch t.Face {
	case 0:
		x, y = t.X+w-dx, t.Y-2*dy
	case 1:
		x, y = t.X+h*2, t.Y+w-dx
	case 2:
		x, y = t.X+w-dx, t.Y+h*2
	default:
		x, y = t.X-h, t.Y+w-dx
	}
	bullet := &Bullet{
		Image:       img,
		X:           x,
		Y:           y,
		F:           t.Face,
		SpeedFactor: float64(2),
		Owner:       t,
	}
	return bullet
}

type EnemyType int

func NewEnemy(t EnemyType, F int, x, y float64) *Tank {
	var img *ebiten.Image
	switch t {
	case 0:
		img = tilesImage.SubImage(image.Rect(32, 0, 32+13, 15)).(*ebiten.Image)
	case 1:
		img = tilesImage.SubImage(image.Rect(48, 0, 48+13, 15)).(*ebiten.Image)
	case 2:
		img = tilesImage.SubImage(image.Rect(64, 0, 64+13, 15)).(*ebiten.Image)
	case 3:
		img = tilesImage.SubImage(image.Rect(80, 0, 80+13, 15)).(*ebiten.Image)
	case 4:
		img = tilesImage.SubImage(image.Rect(32, 16, 32+13, 16+15)).(*ebiten.Image)
	case 5:
		img = tilesImage.SubImage(image.Rect(48, 16, 48+13, 16+15)).(*ebiten.Image)
	case 6:
		img = tilesImage.SubImage(image.Rect(64, 16, 64+13, 16+15)).(*ebiten.Image)
	case 7:
		img = tilesImage.SubImage(image.Rect(80, 16, 80+13, 16+15)).(*ebiten.Image)
	}

	tank := &Tank{
		Image:       img,
		X:           x,
		Y:           y,
		Face:        F,
		Failed:      false,
		enemy:       true,
		SpeedFactor: 0.5,
	}
	return tank
}

func NewOther(x, y float64, t int) *Other {
	var img *ebiten.Image
	switch t {
	case 0:
		img = tilesImage.SubImage(image.Rect(56, 64, 56+tileSize/2, 64+tileSize/2)).(*ebiten.Image)
	case 1:
		img = tilesImage.SubImage(image.Rect(48, 72, 48+tileSize/2, 72+tileSize/2)).(*ebiten.Image)
	case 2:
		img = tilesImage.SubImage(image.Rect(64, 64, 64+tileSize/2, 64+tileSize/2)).(*ebiten.Image)
	default:
		img = tilesImage.SubImage(image.Rect(56, 72, 56+tileSize/2, 72+tileSize/2)).(*ebiten.Image)
	}
	o := &Other{
		Image: img,
		X:     x,
		Y:     y,
		T:     t,
	}
	return o
}

func NewPlayer(player int) *Tank {
	var img *ebiten.Image
	var x, y float64
	if player == 0 {
		img = tilesImage.SubImage(image.Rect(0, 0, tileSize-3, tileSize-3)).(*ebiten.Image)
		x, y = 144, 384
	} else {
		img = tilesImage.SubImage(image.Rect(16, 0, 16+tileSize-3, tileSize-3)).(*ebiten.Image)
		x, y = 240+3, 384
	}

	tank := &Tank{
		enemy:       false,
		Image:       img,
		X:           x,
		Y:           y,
		Face:        0,
		SpeedFactor: 1,
	}
	return tank
}

func GetLevel(i int) []string {
	file := fmt.Sprintf("levels/%d", i)
	data, err := levels.Levels.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	return strings.Split(string(data), "\n")
}

type Castle struct {
	Image          *ebiten.Image
	DestroyedImage *ebiten.Image
	mode           Mode //standing: 0, exploding: 1, destroyed: 2
}

func NewCastle() *Castle {
	var img, destroyedImage *ebiten.Image
	img = tilesImage.SubImage(image.Rect(0, 15, tileSize, 15+tileSize)).(*ebiten.Image)
	destroyedImage = tilesImage.SubImage(image.Rect(0, 15, tileSize, 15+tileSize)).(*ebiten.Image)
	castle := &Castle{
		Image:          img,
		DestroyedImage: destroyedImage,
		mode:           0,
	}
	return castle
}

type Game struct {
	mode      Mode
	p0        *Tank
	p1        *Tank
	twoPlayer bool
	bullets   map[*Bullet]struct{}
	// players   map[*Tank]struct{}
	enemys       map[*Tank]struct{}
	castle       *Castle
	level        int
	mapLevel     []string
	others       map[*Other]struct{}
	enemies_left []int
	idx          int
	// audioContext *audio.Context
	// jumpPlayer   *audio.Player
	// hitPlayer    *audio.Player
}

func (g *Game) addPlayer() {
	g.p0 = NewPlayer(0)
	if g.twoPlayer {
		g.p1 = NewPlayer(1)
	}
}

func (g *Game) drawTank(t *Tank, screen *ebiten.Image) {
	if t == nil || t.Failed {
		return
	}
	op := &ebiten.DrawImageOptions{}
	img := t.Image
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	op.GeoM.Translate(float64(-w)/2, float64(-h)/2)
	angle := float64(t.Face) * (math.Pi / 2)
	op.GeoM.Rotate(angle)
	op.GeoM.Translate(float64(w)/2.0, float64(h)/2.0)
	op.GeoM.Scale(2, 2)
	op.GeoM.Translate(t.X, t.Y)
	screen.DrawImage(img, op)
	width, height, x, y := t.GetInfo()
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(width), float32(height), color.RGBA{255, 0, 0, 20}, true)
}

func (g *Game) DrawCastle(screen *ebiten.Image) {
	c := g.castle
	var img *ebiten.Image
	if c.mode == 0 {
		img = c.Image
	} else if c.mode == 2 {
		img = c.DestroyedImage
	}
	op := &ebiten.DrawImageOptions{}
	// w, h := img.Bounds().Dx(), img.Bounds().Dy()
	// op.GeoM.Translate(float64(w)/2, float64(h)/2)
	op.GeoM.Scale(2, 2)
	op.GeoM.Translate(192, 384)
	screen.DrawImage(img, op)
}

func (g *Game) DrawOther(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	for other := range g.others {
		op.GeoM.Reset()
		op.GeoM.Scale(2, 2)
		op.GeoM.Translate(other.X, other.Y)
		screen.DrawImage(other.Image, op)
		width, height, x, y := other.GetInfo()
		vector.DrawFilledRect(screen, float32(x), float32(y), float32(width), float32(height), color.RGBA{255, 0, 0, 30}, true)
	}
}

func (g *Game) Move(t *Tank) bool {
	// var x, y float64
	if t.Failed {
		return false
	}
	x0, y0 := t.X, t.Y
	switch t.Face {
	case 0:
		t.Y -= t.SpeedFactor
	case 1:
		t.X += t.SpeedFactor
	case 2:
		t.Y += t.SpeedFactor
	default:
		t.X -= t.SpeedFactor
	}

	if t != g.p0 && !g.p0.Failed {
		if CheckCollision(g.p0, t, true) {
			t.X, t.Y = x0, y0
			return true
		}
	}
	if g.p1 != nil && t != g.p1 && !g.p1.Failed {
		if CheckCollision(g.p1, t, true) {
			t.X, t.Y = x0, y0
			return true
		}
	}

	for e := range g.enemys {
		if e != t && CheckCollision(e, t, true) {
			t.X, t.Y = x0, y0
			return true
		}
	}

	for other := range g.others {
		if CheckCollision(t, other, true) {
			t.X, t.Y = x0, y0
			return true
		}
	}

	if g.OutOfScreen(t) {
		t.X, t.Y = x0, y0
		return true
	}
	return false

}

func (g *Game) EnemyMove(t *Tank) {
	// var x, y float64
	if t.Failed {
		return
	}
	switch t.Face {
	case 0:
		t.Y -= t.SpeedFactor
	case 1:
		t.X += t.SpeedFactor
	case 2:
		t.Y += t.SpeedFactor
	default:
		t.X -= t.SpeedFactor
	}
}

func (g *Game) HitAndRemove(b *Bullet) {
	if !g.p0.Failed && b.Owner != g.p0 {
		if CheckCollision(g.p0, b, false) {
			g.p0.Failed = true
			delete(g.bullets, b)
			return
		}
	}
	if g.p1 != nil && !g.p1.Failed && b.Owner != g.p1 {
		if CheckCollision(g.p1, b, false) {
			g.p1.Failed = true
			delete(g.bullets, b)
			return
		}
	}
	for other := range g.others {
		if CheckCollision(other, b, false) {
			delete(g.others, other)
			delete(g.bullets, b)
			return
		}
	}
	for e := range g.enemys {
		if b.Owner != e && CheckCollision(e, b, false) {
			delete(g.enemys, e)
			delete(g.bullets, b)
			return
		}
	}
}

func (g *Game) DrawBullet(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	for bullet := range g.bullets {
		op.GeoM.Reset()
		img := bullet.Image
		w, h := img.Bounds().Dx(), img.Bounds().Dy()
		op.GeoM.Translate(float64(-w)/2, float64(-h)/2)
		angle := float64(bullet.F) * (math.Pi / 2)
		op.GeoM.Rotate(angle)
		op.GeoM.Translate(float64(w)/2, float64(h)/2)
		op.GeoM.Scale(2, 2)
		op.GeoM.Translate(bullet.X, bullet.Y)
		screen.DrawImage(img, op)
		// ebitenutil.DrawRect(screen, 100, 100, 200, 100, color.RGBA{255, 0, 0, 255})
		width, height, x, y := bullet.GetInfo()
		vector.DrawFilledRect(screen, float32(x), float32(y), float32(width), float32(height), color.RGBA{255, 0, 0, 255}, false)
	}
	// w, h := img.Bounds().Dx(), img.Bounds().Dy()
	// op.GeoM.Translate(float64(w)/2, float64(h)/2)
}

func (g *Game) addBullet(bullet *Bullet) {
	g.bullets[bullet] = struct{}{}
}

func (g *Game) addOther(o *Other) {
	g.others[o] = struct{}{}
}

func (g *Game) addEnermy(enemy *Tank) {
	g.enemys[enemy] = struct{}{}
}

func NotSafe(t1, t2 *Tank) bool {
	w1, h1, x_1, y_1 := t1.GetInfo()
	x1, y1 := x_1+float64(w1)/2, y_1+float64(h1)/2
	w2, h2, x_2, y_2 := t2.GetInfo()
	x2, y2 := x_2+float64(w2)/2, y_2+float64(h2)/2
	if math.Pow(x1-x2, 2)+math.Pow(y1-y2, 2) <= math.Pow(float64(w1), 2)+math.Pow(float64(h1), 2) {
		return true
	}
	return false

}

func (g *Game) PosConflict(enemy *Tank) bool {
	if !g.p0.Failed && NotSafe(g.p0, enemy) {
		return true
	}

	if g.p1 != nil && !g.p1.Failed && NotSafe(g.p0, enemy) {
		return true
	}

	for e := range g.enemys {
		if NotSafe(e, enemy) {
			return true
		}
	}
	return false

}

func (g *Game) Generate_enemy() {
	if g.idx < 20 {
		// f := rand.Intn(4)
		pos := born_positions[rand.Intn(3)]
		x, y := float64(pos[0]), float64(pos[1])
		e := NewEnemy(EnemyType(g.enemies_left[g.idx]), rand.Intn(4), x, y)
		if g.PosConflict(e) || len(g.enemys) == max_enemies {
			return
		}
		g.addEnermy(e)
		g.idx++
	}
}

func (g *Game) OutOfScreen(e Entity) bool {
	w, h, x, y := e.GetInfo()
	x1, y1 := float64(w)+x, float64(h)+y
	if x < 0 || y < 0 {
		return true
	}
	if x1 > screenWidth || y1 > screenHeight {
		return true
	}
	return false

}

func (g *Game) init() {
	// g.players = make(map[*Tank]struct{})
	g.enemys = make(map[*Tank]struct{})
	g.bullets = make(map[*Bullet]struct{})
	g.others = make(map[*Other]struct{})
	g.enemies_left = []int{}
	for i, v := range levels_enemies[g.level-1] {
		for k := 0; k < v; k++ {
			g.enemies_left = append(g.enemies_left, i)
		}
	}
	rand.Shuffle(len(g.enemies_left), func(i, j int) {
		g.enemies_left[i], g.enemies_left[j] = g.enemies_left[j], g.enemies_left[i]
	})
	g.idx = 0
	g.addPlayer()
	g.mapLevel = GetLevel(g.level)
	g.ParseLevel()
}

func NewGame() *Game {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Battle City")

	game := &Game{
		mode:      ModeTitle,
		twoPlayer: false,
		castle:    NewCastle(),
		level:     1,
	}
	return game
}

type Caption struct {
	text string
	x    int
	y    int
}

func (g *Game) ParseLevel() {
	for i, s := range g.mapLevel {
		for j := 0; j < len(s); j++ {
			ch := s[j]
			x, y := float64(j*tileSize), float64(i*tileSize)
			var o *Other
			switch ch {
			case '#': //brick
				o = NewOther(x, y, 0)
			case '@': //steel
				o = NewOther(x, y, 1)
			case '%': //water
				o = NewOther(x, y, 2)
			case '~': //grass
				o = NewOther(x, y, 3)
			default:
			}
			if ch != '.' {
				g.addOther(o)
			}
		}
	}
}

func (g *Game) DrawIntroScreen(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0x0, 0x0, 0x0, 0xff})
	cap1 := Caption{
		text: string("HI-"),
		x:    170 / 2,
		y:    135 / 2,
	}
	cap2 := Caption{
		text: string("1 PLAYER"),
		x:    165 / 2,
		y:    250 / 2,
	}
	cap3 := Caption{
		text: string("2 PLAYERS"),
		x:    165 / 2,
		y:    275 / 2,
	}
	cap4 := Caption{
		text: string("(c) 1980 1985 NAMCO LTD."),
		x:    50 / 2,
		y:    350 / 2,
	}
	cap5 := Caption{
		text: string("ALL RIGHTS RESERVED"),
		x:    85 / 2,
		y:    380 / 2,
	}
	var captions = [5]Caption{cap1, cap2, cap3, cap4, cap5}
	for _, cap := range captions {
		text.Draw(screen, cap.text, smallArcadeFont, cap.x, cap.y, color.White)
	}

}

func (g *Game) GetDirection(t *Tank) int {
	var directions [4]int
	cur_dir := t.Face
	dir0 := (t.Face + 1) % 4
	oppo_dir := (t.Face + 2) % 4
	dir1 := (t.Face + 3) % 4

	r := rand.Intn(2)
	if r == 0 {
		directions = [4]int{cur_dir, dir0, dir1, oppo_dir}
	} else {
		directions = [4]int{cur_dir, dir1, dir0, oppo_dir}
	}

	w0, h0, x0, y0 := t.GetInfo()
	var x, y float64
	var collid bool
	var W, H int
	var X, Y float64
	for _, dir := range directions {
		switch dir {
		case 0:
			y = y0 - t.SpeedFactor
		case 1:
			x = x0 + t.SpeedFactor
		case 2:
			y = y0 + t.SpeedFactor
		case 3:
			x = x0 - t.SpeedFactor
		}

		collid = false
		if t != g.p0 && !g.p0.Failed {
			W, H, X, Y = g.p0.GetInfo()
			if RectCollision(W, H, X, Y, w0, h0, x, y, true) {
				collid = true
			}
		}
		if g.p1 != nil && t != g.p1 && !g.p1.Failed {
			W, H, X, Y = g.p1.GetInfo()
			if RectCollision(W, H, X, Y, w0, h0, x, y, true) {
				collid = true
			}
		}

		for e := range g.enemys {
			W, H, X, Y = e.GetInfo()
			if e != t && RectCollision(W, H, X, Y, w0, h0, x, y, true) {
				collid = true
			}
		}

		for other := range g.others {
			W, H, X, Y = other.GetInfo()
			if RectCollision(W, H, X, Y, w0, h0, x, y, true) {
				collid = true
			}
		}

		if x < 0 || y < 0 || x+float64(w0) > screenWidth || y+float64(h0) > screenHeight {
			collid = true
		}
		if !collid {
			return dir
		}
	}
	return -1
}

func (g *Game) Update() error {
	switch g.mode {
	case ModeTitle:
		if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyDown) {
			g.twoPlayer = !g.twoPlayer
		}
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.mode = ModeGame
			g.init()
		}
	case ModeGame:
		for bullet := range g.bullets {
			if bullet.X <= 0 || bullet.X >= screenWidth || bullet.Y <= 0 || bullet.Y >= screenHeight {
				delete(g.bullets, bullet)
			} else {
				bullet.Move()
				g.HitAndRemove(bullet)
			}
		}

		for e := range g.enemys {
			dir := g.GetDirection(e)
			if dir == -1 {
				e.Face = (e.Face + 2) % 2
			} else if dir == e.Face {
				// g.EnemyMove(e)
				g.Move(e)
			} else {
				e.Face = dir
			}
		}

		g.Generate_enemy()

		if !g.p0.Failed {
			if ebiten.IsKeyPressed(ebiten.KeyW) {
				if g.p0.Face == 0 {
					// g.p0.Y -= g.p0.SpeedFactor
					g.Move(g.p0)
				} else {
					g.p0.Face = 0
				}
			} else if ebiten.IsKeyPressed(ebiten.KeyD) {
				if g.p0.Face == 1 {
					// g.p0.X += g.p0.SpeedFactor
					g.Move(g.p0)
				} else {
					g.p0.Face = 1
				}
			} else if ebiten.IsKeyPressed(ebiten.KeyS) {
				if g.p0.Face == 2 {
					// g.p0.Y += g.p0.SpeedFactor
					g.Move(g.p0)
				} else {
					g.p0.Face = 2
				}
			} else if ebiten.IsKeyPressed(ebiten.KeyA) {
				if g.p0.Face == 3 {
					// g.p0.X -= g.p0.SpeedFactor
					g.Move(g.p0)
				} else {
					g.p0.Face = 3
				}
			} else if inpututil.IsKeyJustPressed(ebiten.KeyF) {
				b := g.p0.Fire()
				g.addBullet(b)
			}
		}
		if g.p1 != nil && !g.p1.Failed {
			if ebiten.IsKeyPressed(ebiten.KeyUp) {
				if g.p1.Face == 0 {
					// g.p1.Y -= g.p1.SpeedFactor
					g.Move(g.p1)
				} else {
					g.p1.Face = 0
				}
			} else if ebiten.IsKeyPressed(ebiten.KeyRight) {
				if g.p1.Face == 1 {
					// g.p1.X += g.p1.SpeedFactor
					g.Move(g.p1)
				} else {
					g.p1.Face = 1
				}
			} else if ebiten.IsKeyPressed(ebiten.KeyDown) {
				if g.p1.Face == 2 {
					// g.p1.Y += g.p1.SpeedFactor
					g.Move(g.p1)
				} else {
					g.p1.Face = 2
				}
			} else if ebiten.IsKeyPressed(ebiten.KeyLeft) {
				if g.p1.Face == 3 {
					// g.p1.X -= g.p1.SpeedFactor
					g.Move(g.p1)
				} else {
					g.p1.Face = 3
				}
			} else if inpututil.IsKeyJustPressed(ebiten.KeyControlRight) {
				b := g.p1.Fire()
				g.addBullet(b)
			}
		}

	case ModeGameOver:
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.mode {
	case ModeTitle:
		g.DrawIntroScreen(screen)

	case ModeGame:
		op := &ebiten.DrawImageOptions{}
		for brick := range g.others {
			op.GeoM.Reset()
			op.GeoM.Scale(2, 2)
			op.GeoM.Translate(brick.X, brick.Y)
			screen.DrawImage(brick.Image, op)
		}
		g.drawTank(g.p0, screen)
		g.drawTank(g.p1, screen)
		for e := range g.enemys {
			g.drawTank(e, screen)
		}
		g.DrawCastle(screen)
		g.DrawBullet(screen)
		g.DrawOther(screen)

	case ModeGameOver:
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	g := NewGame()

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
