package main


import (
	"fmt"
	"time"
	"github.com/LDCS/sflag"
	"github.com/LDCS/cim"
	"github.com/LDCS/genutil"
	"strings"
	"os"
	
)



var opt = struct{
	Usage      string     "Updates alert database"
	Kvplist_   string     "Key value pair string. Example : subtab=adm;level=critical;subject=myhost:/disk1 critically full at 95%;escalate=adm;escalate-minutes1=60;escalate-minutes2=180"
	Host_      string     "Host name of alertbase server|localhost"
}{}

func parse_args() {

	sflag.Parse(&opt)

	if opt.Kvplist_ == "" { 
		fmt.Println("monalert: Error in named arguments. See monalert --help")
		os.Exit(1)
	}
}

func parse_kvplist() map[string]string {
	kvp           :=   make(map[string]string)
	opt.Kvplist_   =   strings.Trim(opt.Kvplist_, " \t\n\r;")
	pairs         :=   strings.Split(opt.Kvplist_, ";")

	for _, pair := range pairs {
		tmp    :=  strings.Split(pair, "=")
		kvp[tmp[0]] = tmp[1]
	}
	if _, ok := kvp["subtab"];           ok == false { fmt.Println("No subtab key in kvplist");           os.Exit(1) }
	if _, ok := kvp["level"];            ok == false { fmt.Println("No level key in kvplist");            os.Exit(1) }
	if _, ok := kvp["subject"];          ok == false { fmt.Println("No subject key in kvplist");          os.Exit(1) }
	if _, ok := kvp["escalate"];         ok == false { fmt.Println("No escalate key in kvplist");         os.Exit(1) }
	if _, ok := kvp["escalate-minutes1"];ok == false { fmt.Println("No escalate-minutes1 key in kvplist");os.Exit(1) }
	if _, ok := kvp["escalate-minutes2"];ok == false { fmt.Println("No escalate-minutes2 key in kvplist");os.Exit(1) }

	kvp["subject"]              =     strings.Replace(kvp["subject"], ",", ".", -1)
	kvp["subjectnum"]           =     "0"
	kvp["doneat"]               =     "0"

	if _, ok := kvp["openat"]; ok == false {
		
		openat                 :=     time.Now()
		kvp["openat"]           =     fmt.Sprintf("%d", int(openat.UnixNano()/1000000))
	}

	if _, ok := kvp["owner"];  ok == false {
		kvp["owner"]            =     ""
	}

	kvp["assigner"]             =     ""

	if _, ok :=kvp["status"];  ok == false {
		kvp["status"]           =     "open"
	}

	kvp["comment"]              =     ""
	
	return kvp
	
}


func init_connection() (*cim.CimConnection, error) {
	conn, err := cim.NewCimConnection(opt.Host_, "alertbase01", fmt.Sprintf("alertbase01-%d", time.Now().UnixNano()) )
	return conn, err
}


func run_update(conn *cim.CimConnection, kvp map[string]string) error {

	res, err := conn.RunCommand("editalert " + genutil.GenKVFromMap(kvp))
	if err != nil { return err }
	if res != "done" {
		return fmt.Errorf("alertbase did not return 'done'")
	}
	return nil
}

func main() {
	
	parse_args()
	kvp := parse_kvplist()

	conn, err := init_connection()
	if err != nil {
		fmt.Println("monalert: Error :", err.Error())
		return
	}
	defer conn.Close()
	err = run_update(conn, kvp)
	if err != nil {
		fmt.Println("monalert: Error run_update :", err.Error())
	}
}
