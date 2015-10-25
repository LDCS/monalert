package main


import (
    "fmt"
    "database/sql"
    "time"
    "github.com/LDCS/sflag"
    "strings"
    "os"
    _ "github.com/jbarham/gopgsqldriver"
)



var opt = struct{
	Usage      string     "Updates alert database"
	Kvplist_   string     "Key value pair string. Example : subtab=adm;level=critical;subject=myhost:/disk1 critically full at 95%;escalate=adm;escalate-minutes1=60;escalate-minutes2=180"
}{}

func parse_args() {

	sflag.Parse(&opt)

    if opt.Kvplist_ == "" { 
		fmt.Println("monalert: Error in named arguments. See jobstat --help")
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
    kvp["doneat"]               =     "0001-01-01 00:00:00"

    if _, ok := kvp["openat"]; ok == false {
		
		openat                 :=     time.Now()
		kvp["openat"]           =     openat.Format("2006-01-02 15:04:05")
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


func init_connection(host string) (*sql.DB, error) {
    db, err := sql.Open("postgres", "dbname=ldcs user=ops port=10080 host=" + host + " sslmode=disable")
    return db, err
}


func call_stored_proc(db *sql.DB, kvp map[string]string) error {

    stmt, err := db.Prepare("select ldcs_misc.alertUpSert( $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13 )")
    if err != nil {
        return err
    }
    _, err = stmt.Exec(kvp["subtab"], kvp["subject"], kvp["subjectnum"], kvp["level"], kvp["openat"], kvp["doneat"], kvp["owner"], kvp["assigner"],
		kvp["status"], kvp["comment"], kvp["escalate"], kvp["escalate-minutes1"], kvp["escalate-minutes2"])
    // Unknown issue with the db driver
	if err == nil { return nil }
    if err.Error() == `strconv.ParseInt: parsing "": invalid syntax` {
		return nil
    }
    return err

}

func main() {
    
    parse_args()
    kvp := parse_kvplist()

    for _, dbbox := range []string{ "ldcs1" } {
		db, err := init_connection(dbbox)
		if err != nil {
			fmt.Println("monalert: Error :", err.Error())
			continue
		}
		err = call_stored_proc(db, kvp)
		if err != nil {
			fmt.Println("monalert: Error call_stored_proc :", err.Error())
			continue
		}
		db.Close()
    }
    
}
