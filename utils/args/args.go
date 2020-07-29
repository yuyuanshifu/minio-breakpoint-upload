package args

import (
	"strings"
	"fmt"
)

func ParseArgs(args []string) map[string]string {
	  result := make(map[string]string)
	  i:= 1 
	  var last_key string = ""

	  for i < len(args){
		  key:= args[i]
		  if 0 == strings.Index(key,"-"){
			 if "" != last_key{
				 result[last_key] = "EXIST_OPTION"
			 }

			 last_key = strings.Join(strings.Split(key,"-"),"")
		  } else {
             if last_key != ""{
				fmt.Println(last_key,key)
                result[last_key] = key
			 }
			 last_key = ""
		  }
		  i= i+1
	  }

	  return result
}