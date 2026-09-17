package main

import "fmt"

// twoSum находит индексы двух чисел, которые в сумме дают target
func twoSum(nums []int, target int) []int {
        m := make(map[int]int)
        for i, num := range nums {
               if idx, found := m[target-num]; found {
                       return []int{idx, i}
               }
               m[num] = i
        }
        return nil
}

func main() {
        // Вызываем функцию twoSum с любыми тестовыми аргументами
        testNums := []int{2,7,11,15}
        testTarget := 9

        result := twoSum(testNums, testTarget)
 
        fmt.Printf("Результат выполнения Two Sum: %v\n", result)
}