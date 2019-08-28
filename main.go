package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

// Difficulty 困难度
type Difficulty int

func (d Difficulty) String() string {
	var str string
	switch d {
	case DifficultyEasy:
		str = "Easy"
	case DifficultyMedium:
		str = "Medium"
	case DifficultyHard:
		str = "Hard"
	}
	return str
}

const (
	DifficultyEasy Difficulty = iota + 1
	DifficultyMedium
	DifficultyHard
)

var (
	// BoardConfig 不同困难度的相关配置
	BoardConfig = map[Difficulty]Config{
		DifficultyEasy:   {9, 9, 10},
		DifficultyMedium: {16, 16, 40},
		DifficultyHard:   {30, 16, 99},
	}
)

// Config 游戏配置
type Config struct {
	Width     int
	Height    int
	MineCount int
}

func main() {
	// 选择困难度
	var difficulty Difficulty
	for {
		fmt.Println("Select difficulty: 1-EASY 2-MEDIUM 3-HARD")
		_, err := fmt.Scanln(&difficulty)
		if err != nil || difficulty <= 0 || difficulty > 3 {
			fmt.Println("Wrong input! Please re-select.")
			continue
		}
		fmt.Println("Difficulty:", difficulty.String())
		break
	}
	// 清除屏幕内容
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stderr
	cmd.Run()
	// 初始化游戏
	board := NewBoard(difficulty)
	var input string
	// 正则表达式匹配输入字符串
	reg := regexp.MustCompile(`^(\d{1,2}),(\d{1,2}),([mcs])`)
	// 记录已经扫过的坐标，避免重复输入
	coordRecords := make(map[string]interface{})
	for {
		board.Display()
		fmt.Println("Input coordinates: eg. `0,1,c`, `2,1,m` or `3,3,s`")
		_, err := fmt.Scanln(&input)
		// 清除屏幕内容
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stderr
		cmd.Run()
		fmt.Printf("Your input is %s\n", input)
		if err != nil {
			fmt.Println("Wrong input! Please try again.")
			continue
		}
		if !reg.MatchString(input) {
			fmt.Println("Input format is incorrect! Please try again.")
			continue
		}
		parts := reg.FindStringSubmatch(input)
		x, _ := strconv.Atoi(parts[1])
		y, _ := strconv.Atoi(parts[2])
		coord := Coordinate{X: x, Y: y}
		if !board.Contains(coord) {
			fmt.Println("Wrong coordinates! Please try again.")
			continue
		}
		if _, ok := coordRecords[coord.String()]; ok {
			fmt.Println("The mine has been cleaned! Please re-input another coordinates.")
			continue
		}
		mine := &board.Rows[coord.Y][coord.X]
		// 进行扫雷操作
		if parts[3] == "c" {
			mine.Status = StatusCleaned
			if mine.HasMine {
				mine.Status = StatusBombed
				board.Display()
				fmt.Printf("You lose! Score is %d\n", board.Score)
				break
			}
			coordRecords[coord.String()] = nil
			board.Score += 1
			board.SweptCount += 1
			board.Clean(coord)
			// 进行标记操作
		} else if parts[3] == "m" {
			mine.Status = StatusMarked
			if mine.HasMine {
				board.SweptCount += 1
				board.MineSweptCount += 1
				board.Score += 2
			}
			// 进行判疑操作
		} else {
			mine.Status = StatusSuspicious
		}
		if board.Debug {
			fmt.Println("SweptCount:", board.SweptCount)
			fmt.Println("MineSweptCount:", board.MineSweptCount)
		}
		// 判断是否胜利
		if board.IsWin() {
			board.Display()
			fmt.Printf("You win! Score is %d\n", board.Score)
			return
		}
	}
}
