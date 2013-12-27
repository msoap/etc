package main

import (
    "fmt"
    "os"
    "os/exec"
    "path"
    "strconv"
    "strings"
)

const NUMS = 1500

//------------------------------------
func test_string_key(nums int) interface{} {
    fmt.Println("Start")

    var hash = make(map[string]map[string]string, nums)

    for i := 0; i < nums; i++ {
        str_i := fmt.Sprintf("%d", i)
        hash["key_"+str_i] = make(map[string]string, nums)

        for j := 0; j < nums; j++ {
            str_j := fmt.Sprintf("%d", j)
            hash["key_"+str_i]["key_j_"+str_j] = "string_" + str_i + "_" + str_j
        }
    }

    fmt.Println("Finish")
    return hash
}

//------------------------------------
func test_int_key(nums int) interface{} {
    fmt.Println("Start")

    var hash = make(map[int]map[int]string, nums)

    for i := 0; i < nums; i++ {
        hash[i] = make(map[int]string, nums)

        for j := 0; j < nums; j++ {
            hash[i][j] = "string_" + fmt.Sprintf("%d", i) + "_" + fmt.Sprintf("%d", j)
        }
    }

    fmt.Println("Finish")
    return hash
}

//------------------------------------
func test_slice_of_string(nums int) interface{} {
    fmt.Println("Start slice of string")

    var array = make([][]string, nums)

    for i := 0; i < nums; i++ {
        array[i] = make([]string, nums)

        for j := 0; j < nums; j++ {
            array[i][j] = "string_" + fmt.Sprintf("%d", i) + "_" + fmt.Sprintf("%d", j)
        }
    }

    fmt.Println("Finish")
    return array
}

//------------------------------------
func test_slice_of_int(nums int) interface{} {
    fmt.Println("Start slice of int")

    var array = make([][]int, nums)

    for i := 0; i < nums; i++ {
        array[i] = make([]int, nums)

        for j := 0; j < nums; j++ {
            array[i][j] = i * nums + j
        }
    }

    fmt.Println("Finish")
    return array
}

//------------------------------------
func test_array(nums int) interface{} {
    fmt.Println("Start array")

    var array [NUMS][NUMS]string

    for i := 0; i < nums; i++ {
        // array[i] = make([]string, nums)

        for j := 0; j < nums; j++ {
            array[i][j] = "string_" + fmt.Sprintf("%d", i) + "_" + fmt.Sprintf("%d", j)
        }
    }

    fmt.Println("Finish")
    return array
}

//------------------------------------
func keys(hash map[string]func(int)interface{}) []string {
    result := []string{}

    for key, _ := range hash {
        result = append(result, key)
    }

    return result
}

//------------------------------------
func main() {
    test_type := ""
    if len(os.Args) > 1 {
        test_type = os.Args[1]
    }

    nums := NUMS
    if len(os.Args) == 3 {
        nums_arg, err := strconv.Atoi(os.Args[2])
        if err == nil && nums_arg > 0 {
            nums = nums_arg
        }
    }
    fmt.Printf("nums: %d²\n", nums)

    test_functions := map[string]func(int)interface{}{"test_int_key": test_int_key,
                                                      "test_string_key": test_string_key,
                                                      "test_slice_of_string": test_slice_of_string,
                                                      "test_slice_of_int": test_slice_of_int,
                                                      "test_array": test_array}

    var result interface{}
    if test_function, found := test_functions[test_type]; found {
        result = test_function(nums)
        fmt.Printf("type is: %T\n", result)
    } else {
        fmt.Printf("usage: %s %s [NUMS]\n", path.Base(os.Args[0]), strings.Join(keys(test_functions), "|"))
        return
    }

    cmd_out, err := exec.Command("sh", "-c", "ps aux | awk '$2 == " + strconv.Itoa(os.Getpid()) + " {print $6}'").Output()
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    memory_kb, _ := strconv.Atoi(strings.TrimSpace(string(cmd_out)))
    fmt.Printf("memory: %.2f MB\n", float32(memory_kb) / 1024.0)
}
