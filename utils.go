package ultron

import (
	"fmt"
)

func showLogo() {
	fmt.Println(`
      ___           ___       ___           ___           ___           ___     
     /\__\         /\__\     /\  \         /\  \         /\  \         /\__\    
    /:/  /        /:/  /     \:\  \       /::\  \       /::\  \       /::|  |   
   /:/  /        /:/  /       \:\  \     /:/\:\  \     /:/\:\  \     /:|:|  |   
  /:/  /  ___   /:/  /        /::\  \   /::\~\:\  \   /:/  \:\  \   /:/|:|  |__ 
 /:/__/  /\__\ /:/__/        /:/\:\__\ /:/\:\ \:\__\ /:/__/ \:\__\ /:/ |:| /\__\
 \:\  \ /:/  / \:\  \       /:/  \/__/ \/_|::\/:/  / \:\  \ /:/  / \/__|:|/:/  /
  \:\  /:/  /   \:\  \     /:/  /         |:|::/  /   \:\  /:/  /      |:/:/  / 
   \:\/:/  /     \:\  \    \/__/          |:|\/__/     \:\/:/  /       |::/  /  
    \::/  /       \:\__\                  |:|  |        \::/  /        /:/  /   
     \/__/         \/__/                   \|__|         \/__/         \/__/`)
}

func init() {
	showLogo()
}
