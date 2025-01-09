package cmd

import (
	"fmt"
	"runtime"
)

const logo = `
                                                   
 █████ █████  ██████   ████████    ███████  ██████ 
░░███ ░░███  ░░░░░███ ░░███░░███  ███░░███ ███░░███
 ░░░█████░    ███████  ░███ ░███ ░███ ░███░███ ░███
  ███░░░███  ███░░███  ░███ ░███ ░███ ░███░███ ░███
 █████ █████░░████████ ████ █████░░███████░░██████ 
░░░░░ ░░░░░  ░░░░░░░░ ░░░░ ░░░░░  ░░░░░███ ░░░░░░  
                                  ███ ░███         
                                 ░░██████          
                                  ░░░░░░                                                                                         
`

func Welcome() {
	fmt.Println(logo)
	fmt.Println(fmt.Sprintf("Server      Name:      %s", "xan-go"))
	fmt.Println(fmt.Sprintf("System      Name:      %s", runtime.GOOS))
	fmt.Println(fmt.Sprintf("Go          Version:   %s", runtime.Version()[2:]))
}
