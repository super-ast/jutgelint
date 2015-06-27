package main

func main() {
	var x int // x is declared too far from its use, x is defined but its current value isn't used
	if true {
		x = 3 // x is defined but its current value isn't used
		x = 4 // x is defined but its current value isn't used
	}
}
