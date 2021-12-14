package main

import (
	"fmt"
	"nacos/gitnacos"
	"time"
)

func test(dataid, group string) {
	d := gitnacos.ListenConfig(dataid, group)
	for {
		fmt.Println(<-d)
	}

}

func main() {
	gitnacos.GetString("test", "test.group", func(data *string, err error) {
		if err == nil {
			fmt.Println("data:", *data)
		} else {
			fmt.Println(err)
		}
	})

	// fmt.Println(gitnacos.GetConfigList(1, 15))
	gitnacos.GetString("test2", "test.group", func(data *string, err error) {
		if err == nil {
			fmt.Println("data2:", *data)
		} else {
			fmt.Println(err)
		}
	})

	d, _ := gitnacos.GetString("test3", "test.group", nil)
	fmt.Println(d)

	time.Sleep(1 * time.Hour)
}
