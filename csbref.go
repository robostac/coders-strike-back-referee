package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

//GAME CONSTANTS
const podRSQ = 800 * 800
const cpRSQ = 599 * 599
const podCount = 4
const minImpulse = 120
const frictionVal = 0.85
const checkpointGenerationGap = 30

const p2Bug = true //run turns in serial (slower) so the p2 angle bug can be tested

//MATH CONSTANTS
const fullCircle = (2 * math.Pi)
const radToDeg = 180.0 / math.Pi
const degToRad = math.Pi / 180.0
const maxRotate = (18.0 * degToRad)

//types

type distanceSqType float64
type gameMap []point

type point struct {
	x float64
	y float64
}

type object struct {
	p           point
	s           point
	angle       float64
	next        int
	shieldtimer int
	boosted     int
}

type playerMove struct {
	target point
	thrust int
	shield bool
	boost  bool
}

type game [podCount]object

var globalCp [50]point
var globalNumCp int
var playerTimeout [2]int

//taken from AGADE CSB RUNNER ARENA
//https://github.com/Agade09/CSB-Runner-Arena/blob/master/Arena.cpp
var possibleMaps = []gameMap{
	{{12460, 1350}, {10540, 5980}, {3580, 5180}, {13580, 7600}},
	{{3600, 5280}, {13840, 5080}, {10680, 2280}, {8700, 7460}, {7200, 2160}},
	{{4560, 2180}, {7350, 4940}, {3320, 7230}, {14580, 7700}, {10560, 5060}, {13100, 2320}},
	{{5010, 5260}, {11480, 6080}, {9100, 1840}}, {{14660, 1410}, {3450, 7220}, {9420, 7240}, {5970, 4240}},
	{{3640, 4420}, {8000, 7900}, {13300, 5540}, {9560, 1400}},
	{{4100, 7420}, {13500, 2340}, {12940, 7220}, {5640, 2580}},
	{{14520, 7780}, {6320, 4290}, {7800, 860}, {7660, 5970}, {3140, 7540}, {9520, 4380}},
	{{10040, 5970}, {13920, 1940}, {8020, 3260}, {2670, 7020}}, {{7500, 6940}, {6000, 5360}, {11300, 2820}},
	{{4060, 4660}, {13040, 1900}, {6560, 7840}, {7480, 1360}, {12700, 7100}},
	{{3020, 5190}, {6280, 7760}, {14100, 7760}, {13880, 1220}, {10240, 4920}, {6100, 2200}},
	{{10323, 3366}, {11203, 5425}, {7259, 6656}, {5425, 2838}}}
var possibleMapCount = len(possibleMaps)

func (p *point) dot(n point) float64 {
	return p.x*n.x + p.y*n.y
}

func (g *game) nextTurn() {
	t := 0.0
	for t < 1.0 {
		first := 100.0
		cli := 0
		clj := 0
		for i := 0; i < podCount; i++ {
			for j := i + 1; j < podCount; j++ {
				tx := g[i].newCollide(&g[j], podRSQ)
				if (tx > 0) && tx+t < 1.0 && (tx < first) {
					first = tx
					cli = i
					clj = j
				}
			}
		}
		if cli == clj {
			g.forwardTime(1.0 - t)
			t = 1.0
		} else {
			g.forwardTime(first)
			g.bounce(cli, clj)
			t += first
		}
	}
	for i := 0; i < podCount; i++ {
		g[i].endTurn()

	}
	playerTimeout[0]--
	playerTimeout[1]--
}

func (g *game) bounce(p1 int, p2 int) {

	oa := &g[p1]
	ob := &g[p2]

	var m1 float64 = 1
	var m2 float64 = 1
	if oa.shieldtimer == 4 {
		m1 = 10
	}
	if ob.shieldtimer == 4 {
		m2 = 10
	}
	mcoeff := ((m1 + m2) / (m1 * m2))

	nx := (oa.p.x - ob.p.x)
	ny := (oa.p.y - ob.p.y)

	nxnysquare := nx*nx + ny*ny
	dvx := (oa.s.x - ob.s.x)
	dvy := (oa.s.y - ob.s.y)

	product := nx*dvx + ny*dvy

	fx := (nx * product) / (nxnysquare * mcoeff)
	fy := (ny * product) / (nxnysquare * mcoeff)

	icheck := (fx*fx + fy*fy)

	impulse := math.Sqrt(icheck)
	oa.s.x -= fx / m1
	oa.s.y -= fy / m1
	ob.s.x += fx / m2
	ob.s.y += fy / m2
	if impulse < minImpulse {
		fx = fx * minImpulse / impulse
		fy = fy * minImpulse / impulse
	}
	oa.s.x -= fx / m1
	oa.s.y -= fy / m1
	ob.s.x += fx / m2
	ob.s.y += fy / m2
}

func getAngle(start point, end point) float64 {
	d := (distance(start, end))
	dx := (end.x - start.x) / (d)
	dy := (end.y - start.y)
	a := (math.Acos((dx)))
	if dy < 0 {
		a = fullCircle - a
	}
	return a
}

func distance2(p1 point, p2 point) distanceSqType {
	x := distanceSqType(p2.x - p1.x)
	x = x * x
	y := distanceSqType(p2.y - p1.y)
	y = y * y
	return x + y
}

func distance(p1 point, p2 point) float64 {
	return (math.Sqrt(float64(distance2(p1, p2))))
}

func (g *game) forwardTime(t float64) {
	for i := 0; i < podCount; i++ {
		obj := &g[i]
		cp := object{}
		cp.p = globalCp[g[i].next]
		tx := obj.newCollide(&cp, cpRSQ)
		if (tx > 0) && tx < t {
			obj.next = (obj.next + 1) % globalNumCp
			if i < 2 {
				playerTimeout[0] = 100
			} else {
				playerTimeout[1] = 100
			}
		}
		obj.p.x += (obj.s.x * (t))
		obj.p.y += (obj.s.y * (t))
	}
}

func round(x float64) float64 {
	if x > 0 {
		x = (math.Floor((x) + 0.5))
	} else {
		x = (math.Ceil((x) - 0.5))
	}
	return x
}

func (obj *object) newCollide(b *object, rsq float64) float64 {

	p := point{b.p.x - obj.p.x, b.p.y - obj.p.y}
	pLength2 := p.x*p.x + p.y*p.y

	if pLength2 <= rsq {
		return 0
	}

	v := point{(b.s.x - obj.s.x), (b.s.y - obj.s.y)}
	dot := p.dot(v)

	if dot > 0 {
		return -1
	}

	vLength2 := v.x*v.x + v.y*v.y
	disc := dot*dot - vLength2*(pLength2-rsq)

	if disc < 0 {
		return -1
	}

	discdist := (math.Sqrt((disc)))
	t1 := (-dot - discdist) / vLength2
	//t2 := (-dot + discdist) / vLength2
	if t1 >= 0 && t1 < 1 {
		return (t1)
	}
	return -1
}

func (obj *object) applyRotate(rotateAngle float64) {
	if rotateAngle < -maxRotate {
		rotateAngle = -maxRotate
	}
	if rotateAngle > maxRotate {
		rotateAngle = maxRotate
	}
	obj.angle = rotateAngle + obj.angle
	for obj.angle < 0 {
		obj.angle += fullCircle
	}
	for obj.angle > fullCircle {
		obj.angle -= fullCircle
	}
}

func (obj *object) applyThrust(t int) {
	cs, cc := math.Sincos(obj.angle)
	obj.s.x += (cc * float64(t))
	obj.s.y += (cs * float64(t))
}

func (obj *object) applyRotateThrust(a float64, t int) {
	obj.applyRotate(a)
	obj.applyThrust(t)
}

func (obj *object) endTurn() {

	obj.s.x = (math.Trunc((obj.s.x * frictionVal)))
	obj.s.y = (math.Trunc((obj.s.y * frictionVal)))
	obj.p.x = round(obj.p.x)
	obj.p.y = round(obj.p.y)

	if obj.shieldtimer > 0 {
		obj.shieldtimer--
	}
}

func (obj *object) diffAngle(p point) float64 {
	a := getAngle(obj.p, p)
	right := a - obj.angle
	if obj.angle > a {
		right = fullCircle - obj.angle + a
	}
	left := obj.angle - a
	if obj.angle < a {
		left = obj.angle + fullCircle - a
	}
	if right < left {
		return right
	}
	return -left
}

func testMode() {

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &globalNumCp)
	for i := 0; i < globalNumCp; i++ {
		var x, y float64
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &x, &y)
		globalCp[i] = point{x, y}
	}
	var nTest int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &nTest)
	var g game
	for tn := 0; tn < nTest; tn++ {
		for i := 0; i < podCount; i++ {
			var p object
			var angle float64
			scanner.Scan()
			fmt.Sscan(scanner.Text(), &p.p.x, &p.p.y, &p.s.x, &p.s.y, &angle, &p.next, &p.shieldtimer, &p.boosted)
			if angle < 0 {
				angle += 360
			}
			p.angle = (angle) * degToRad
			g[i] = p
		}
		for i := 0; i < podCount; i++ {
			var px, py float64
			var thrust string
			var t int
			scanner.Scan()
			fmt.Sscan(scanner.Text(), &px, &py, &thrust)
			t, err := strconv.Atoi(thrust)
			if err != nil {
				t = 0
				if thrust == "SHIELD" {
					g[i].shieldtimer = 4
				} else if thrust == "BOOST" {
					t = 650
					if g[i].boosted == 0 {
						g[i].boosted = 1
					} else {
						t = 200
					}
				}
			}
			if g[i].shieldtimer > 0 {
				t = 0
			}
			angle := g[i].diffAngle(point{px, py})
			g[i].applyRotate(angle)
			g[i].applyThrust(t)
		}
		g.nextTurn()
		for i := 0; i < podCount; i++ {
			p := &g[i]
			fmt.Printf("%d %d %d %d %f %d %d %d\n", int(p.p.x), int(p.p.y), int(p.s.x), int(p.s.y), p.angle*radToDeg, p.next, p.shieldtimer, p.boosted)
		}
	}
}

var startPointMult = [4]point{{500, -500}, {-500, 500}, {1500, -1500}, {-1500, 1500}}

func initialiseGame(g *game, m gameMap) {
	cp1minus0 := point{}
	cp1minus0.x = m[1].x - m[0].x
	cp1minus0.y = m[1].y - m[0].y
	dd := distance(m[1], m[0])
	cp1minus0.x /= dd
	cp1minus0.y /= dd

	for podN := range g {
		p := &g[podN]
		p.angle = -1 * degToRad
		p.next = 1
		p.p.x = m[0].x + cp1minus0.x*startPointMult[podN].x
		p.p.y = m[0].y + cp1minus0.y*startPointMult[podN].y
	}
}

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "-test" {
			testMode()
			return
		}
	}
	playerTimeout[0] = 100
	playerTimeout[1] = 100
	rand.Seed(time.Now().UTC().UnixNano())
	scanner := bufio.NewScanner(os.Stdin)
	started := false
	var players int

	for started == false {
		scanner.Scan()
		startText := strings.Split(scanner.Text(), " ")
		if startText[0] == "###Start" {
			var err error
			players, err = strconv.Atoi(startText[1])
			if err != nil || players != 2 {
				fmt.Fprintln(os.Stderr, "Error with player count input")
				os.Exit(-1)
			}
			started = true
		} else if startText[0] == "###Seed" {
			v, err := strconv.ParseInt(startText[1], 10, 64)
			fmt.Fprintln(os.Stderr, v)
			if err == nil {
				rand.Seed(v)
			}
		} else {
			fmt.Fprintln(os.Stderr, "Unsupported startup command: ", startText[0])
			os.Exit(0)
		}
	}
	currentMap := possibleMaps[rand.Intn(possibleMapCount)]
	for i, v := range currentMap {
		currentMap[i].x = v.x + float64(rand.Intn(checkpointGenerationGap*2+1)-checkpointGenerationGap)
		currentMap[i].y = v.y + float64(rand.Intn(checkpointGenerationGap*2+1)-checkpointGenerationGap)
	}
	//setup global checkpoints
	laps := 3
	for i := 0; i < 3; i++ {
		for _, v := range currentMap {
			globalCp[globalNumCp] = v
			globalNumCp++
		}
	}
	//add last checkpoint at the end
	globalCp[globalNumCp] = currentMap[0]
	globalNumCp++
	var g game
	initialiseGame(&g, currentMap)
	outputSetup(currentMap, 2, laps)
	for turnCount := 0; turnCount < 500; turnCount++ {
		var moves [4]playerMove
		if p2Bug == false {
			for player := 0; player < players; player++ {
				givePlayerOutput(&g, player, currentMap)
			}
		}
		for player := 0; player < players; player++ {
			if p2Bug {
				givePlayerOutput(&g, player, currentMap)
			}
			theseMoves, valid := getPlayerInput(player, scanner)
			if valid == false {
				lostGame(player)
			}
			for i := range theseMoves {
				move := &theseMoves[i]
				pod := &g[player*2+i]
				if move.boost {
					if pod.boosted == 0 {
						pod.boosted = 1
						move.thrust = 650
					} else {
						move.thrust = 200
					}
				}
				if move.shield {
					pod.shieldtimer = 4
				}
				if pod.shieldtimer > 0 {
					move.thrust = 0
				}
				//do thise here to replicate player 2 bug
				if turnCount == 0 {
					pod.angle = pod.diffAngle(move.target)
				} else {
					pod.applyRotate(pod.diffAngle(move.target))
				}
				moves[player*2+i] = *move
			}
		}
		for podN := range g {
			pod := &g[podN]
			pod.applyThrust(moves[podN].thrust)
		}
		g.nextTurn()
		if playerTimeout[0] <= 0 {
			lostGame(0)
		}
		if playerTimeout[1] <= 0 {
			lostGame(1)
		}
		for podN := range g {
			pod := &g[podN]
			if pod.next == 0 {
				if podN < 2 {
					wonGame(0)
				} else {
					wonGame(1)
				}
			}
		}
	}
	winner := 0
	best := 0.0
	for podN := range g {
		score := float64(g[podN].next * 1000000)
		score -= distance(g[podN].p, globalCp[g[podN].next])
		if score > best {
			best = score
			winner = podN
		}
	}
	if winner < 2 {
		wonGame(0)
	} else {
		wonGame(1)
	}
}

func lostGame(player int) {
	winner := 0
	loser := 1
	if player == winner {
		winner, loser = loser, winner
	}
	fmt.Printf("###End %d %d\n", winner, loser)
	os.Exit(0)
}

func wonGame(player int) {
	winner := 0
	loser := 1
	if player == loser {
		winner, loser = loser, winner
	}
	fmt.Printf("###End %d %d\n", winner, loser)
	os.Exit(0)
}

func getPlayerInput(player int, scanner *bufio.Scanner) ([2]playerMove, bool) {
	pm := [2]playerMove{}
	valid := true
	fmt.Printf("###Output %d 2\n", player)
	for i := range pm {
		if scanner.Scan() == false {
			os.Exit(0)
		}
		var thrust string
		fmt.Sscanf(scanner.Text(), "%f %f %s\n", &pm[i].target.x, &pm[i].target.y, &thrust)

		pm[i].thrust = 0
		switch thrust {
		case "SHIELD":
			pm[i].shield = true
		case "BOOST":
			pm[i].boost = true
		default:
			v, err := strconv.Atoi(thrust)
			if err != nil {
				valid = false
			} else {
				if v > 200 {
					valid = false
				}
				pm[i].thrust = v
			}
		}
	}
	return pm, valid
}

func outputSetup(m gameMap, players int, laps int) {
	for player := 0; player < players; player++ {
		fmt.Printf("###Input %d\n", player)
		fmt.Println(laps)
		fmt.Println(len(m))
		for _, v := range m {
			fmt.Println(v.x, v.y)
		}
	}
}

func givePlayerOutput(g *game, player int, m gameMap) {
	pods := [4]int{0, 1, 2, 3}
	if player == 1 {
		pods = [4]int{2, 3, 0, 1}
	}
	fmt.Printf("###Input %d\n", player)
	for _, podN := range pods {
		p := &g[podN]
		fmt.Printf("%d %d %d %d %d %d\n", int(p.p.x), int(p.p.y), int(p.s.x), int(p.s.y), int(round(p.angle*radToDeg)), p.next%len(m))
	}
}
