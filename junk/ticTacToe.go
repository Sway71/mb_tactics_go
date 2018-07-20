package tactics

import (
	"fmt"
)

func main() {
  var outcome = "unfinished"
  var gameBoard = [][]int{
    {0, 1, 1},
    {2, 1, 2},
    {1, 2, 2},
  }

  fmt.Println(gameBoard)

  for (i := 0; i < 3; i++) {
    fmt.Println()
  }
}
