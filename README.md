# 说明
可读取nacos配置并监听配置变更

# 使用方法
```go
gitnacos.GetString("test2", "test.group", func(data *string, err error) {
		if err == nil {
			fmt.Println("data2:", *data)
		} else {
			fmt.Println(err)
		}
	})

	d, _ := gitnacos.GetString("test3", "test.group", nil)
	fmt.Println(d)
  ```
  # 注意事项
  必须将config文件放置在config文件夹中，config文件夹位于项目根目录。
  
  # config内容
  ```json
  
  ```
