package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/alexeyco/simpletable"
)

// Coordinate 坐标
type Coordinate struct {
	X int
	Y int
}

// GetSurroundingCoordinates 获取某个坐标的周围至多8个坐标，以数组形式返回，仅返回符合条件的坐标
func (c *Coordinate) GetSurroundingCoordinates(b *Board) (coords []Coordinate) {
	for i := -1; i < 2; i++ {
		for j := -1; j < 2; j++ {
			if i == 0 && j == 0 {
				continue
			}
			var coord Coordinate
			x := c.X + i
			y := c.Y + j
			if x < 0 || x >= b.ColCount {
				continue
			}
			if y < 0 || y >= b.RowCount {
				continue
			}
			coord.X = x
			coord.Y = y
			coords = append(coords, coord)
		}
	}
	return
}

func (c Coordinate) String() string {
	return fmt.Sprintf("(%d, %d)", c.X, c.Y)
}

// Status 地雷状态
type Status int

const (
	StatusUnswept    Status = iota // 未扫除
	StatusCleaned                  // 已扫除
	StatusMarked                   // 已标记
	StatusSuspicious               // 已判疑
	StatusBombed                   // 已爆炸
)

// Square 方块
type Square struct {
	Num     int    // 方块周围存在地雷数
	HasMine bool   // 该方块是否有地雷
	Status  Status // 方块状态
}

// Display 根据状态返回显示内容
func (s *Square) Display(debug bool) string {
	var symbol string
	switch s.Status {
	case StatusUnswept:
		symbol = "■"
	case StatusCleaned:
		if s.Num > 0 {
			symbol = strconv.Itoa(s.Num)
		} else {
			symbol = ""
		}
	case StatusMarked:
		symbol = "△"
	case StatusSuspicious:
		symbol = "?"
	case StatusBombed:
		symbol = "@"
	}
	if debug {
		if s.HasMine {
			symbol = "*"
		} else {
			symbol = strconv.Itoa(s.Num)
		}
	}
	return symbol
}

// Row 行
type Row []Square

// Board 游戏板
type Board struct {
	Score          int
	RowCount       int
	ColCount       int
	MineCount      int
	SweptCount     int
	MineSweptCount int
	Rows           []Row
	Debug          bool
}

// NewBoard 创建并初始化新的游戏板
func NewBoard(difficulty Difficulty) *Board {
	config := BoardConfig[difficulty]
	var board Board
	board.ColCount = config.Width
	board.RowCount = config.Height
	board.MineCount = config.MineCount
	board.Init()
	return &board
}

// Init 初始化
func (b *Board) Init() {
	b.Rows = make([]Row, b.RowCount)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// 生成随机地雷坐标数据
	mineCoords := make(map[string]interface{})
	var mineCount int
	for {
		x := r.Intn(b.ColCount)
		y := r.Intn(b.RowCount)
		coord := Coordinate{X: x, Y: y}
		if _, ok := mineCoords[coord.String()]; ok {
			continue
		}
		mineCoords[coord.String()] = nil
		mineCount++
		if mineCount == b.MineCount {
			break
		}
	}
	// 填充地雷
	for i := 0; i < b.RowCount; i++ {
		var row Row
		row = make([]Square, b.ColCount)
		for j := 0; j < b.ColCount; j++ {
			var mine Square
			coord := Coordinate{X: j, Y: i}
			if _, ok := mineCoords[coord.String()]; ok {
				mine.HasMine = true
			}
			for _, c := range coord.GetSurroundingCoordinates(b) {
				if _, ok := mineCoords[c.String()]; ok {
					mine.Num += 1
				}
			}
			row[j] = mine
		}
		b.Rows[i] = row
	}
}

// Display 显示游戏板
func (b *Board) Display() {
	table := simpletable.New()
	// 设置表头
	var header simpletable.Header
	var cells []*simpletable.Cell
	cells = make([]*simpletable.Cell, b.ColCount+1)
	cells[0] = &simpletable.Cell{Align: simpletable.AlignLeft, Text: "Y\\X"}
	for i := 0; i < b.ColCount; i++ {
		var cell simpletable.Cell
		cell.Align = simpletable.AlignCenter
		cell.Text = strconv.Itoa(i)
		cells[i+1] = &cell
	}
	header.Cells = cells
	table.Header = &header
	// 设置数据
	table.Body.Cells = make([][]*simpletable.Cell, 2*b.RowCount-1)
	for i := 0; i < b.RowCount; i++ {
		row := make([]*simpletable.Cell, b.ColCount+1)
		row[0] = &simpletable.Cell{Align: simpletable.AlignCenter, Text: strconv.Itoa(i)}
		for j := 0; j < b.ColCount; j++ {
			var cell simpletable.Cell
			cell.Align = simpletable.AlignCenter
			cell.Text = b.Rows[i][j].Display(b.Debug)
			row[j+1] = &cell
		}
		table.Body.Cells[2*i] = row
		if i != b.RowCount-1 {
			table.Body.Cells[2*i+1] = b.getSeparationLine(b.ColCount + 1)
		}
	}
	table.SetStyle(simpletable.StyleUnicode)
	table.Println()
}

// getSeparationLine 获取水平分割线
func (b *Board) getSeparationLine(length int) (row []*simpletable.Cell) {
	var cell simpletable.Cell
	cell.Align = simpletable.AlignRight
	cell.Text = strings.Repeat("━", (length-1)*5)
	cell.Span = length
	row = append(row, &cell)
	return row
}

// Contains 该游戏板是否包含某个坐标
func (b *Board) Contains(c Coordinate) bool {
	if c.X < 0 || c.X >= b.ColCount || c.Y < 0 || c.Y >= b.RowCount {
		return false
	}
	return true
}

// Clean 开辟安全的区域
func (b *Board) Clean(c Coordinate) {
	coordsSet := make(map[string]interface{})
	b.clean(coordsSet, c)
}

// clean 递归开辟安全的区域
func (b *Board) clean(coordsSet map[string]interface{}, c Coordinate) {
	if _, ok := coordsSet[c.String()]; ok {
		return
	}
	coordsSet[c.String()] = nil
	coords := c.GetSurroundingCoordinates(b)
	var hasBomb bool
	for _, coord := range coords {
		m := b.Rows[coord.Y][coord.X]
		if m.HasMine {
			hasBomb = true
			break
		}
	}
	if !hasBomb {
		for _, coord := range coords {
			m := &b.Rows[coord.Y][coord.X]
			if m.Status != StatusCleaned {
				m.Status = StatusCleaned
				b.SweptCount += 1
				b.Score += 1
			}
			b.clean(coordsSet, coord)
		}
	}
}

// IsWin 是否游戏胜利
func (b *Board) IsWin() bool {
	return b.SweptCount == b.RowCount*b.ColCount && b.MineSweptCount == b.MineCount
}
