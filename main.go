package main

import "fmt"

func twoSum(numbers []int, target int) []int {
	seen := map[int]int{}
	for i, num := range numbers {
		index, ok := seen[target-num]
		if ok == true {
			return []int{i, index}
		}
		seen[num] = i
	}
	return nil
}

func main() {
	nums := []int{2, 7, 11, 15}
	target := 9

	fmt.Println(twoSum(nums, target))
}
