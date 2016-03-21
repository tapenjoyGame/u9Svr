package android

import (
	//"fmt"
	//"github.com/astaxie/beego"
	//"os"
	"os/exec"
)

const (
	apkTool  = "package/tools/apktool_2.0.2.jar"
	signTool = "jarsigner"
)

//apk反编
func UnCompileApk(apkFile string, outPath string) (err error) {
	var args []string
	args = make([]string, 7)
	args[0] = "-jar"
	args[1] = apkTool
	args[2] = "d"
	args[3] = "-f"
	args[4] = apkFile
	args[5] = "-o"
	args[6] = outPath
	cmd := exec.Command("java", args...)

	//var buf []byte
	//buf, err = cmd.Output()
	_, err = cmd.Output()

	//if err != nil {
	//fmt.Fprintf(os.Stderr, "The command failed to perform: %s", err)
	//}
	//fmt.Fprintf(os.Stdout, "Result: %s", buf)
	return
}

// java -jar package/tools/apktool_2.0.2.jar b -f package/uncompile/122/101_1001_1.0.0 -o package/uncompile/122/101_1001_1.0.0.apk
//apk回编
func CompileApk(apkPath string, outApkFile string) (err error) {
	var args []string
	args = make([]string, 7)

	args[0] = "-jar"
	args[1] = apkTool
	args[2] = "b"
	args[3] = "-f"
	args[4] = apkPath
	args[5] = "-o"
	args[6] = outApkFile
	cmd := exec.Command("java", args...)

	//var buf []byte
	//buf, err = cmd.Output()

	_, err = cmd.Output()
	//result = true
	//if err != nil {
	//beego.Trace(buf)
	//fmt.Fprintf(os.Stderr, "The command failed to perform: %s", err)
	//result = false
	//}

	//fmt.Fprintf(os.Stdout, "Result: %s", buf)
	return
}

//var c chan bool

//apk签名
func ApkSign(keyFile string, password string, unsignApk string, signApk string, aliasName string) (err error) {
	var args []string
	args = make([]string, 13)

	args[0] = "-verbose"
	args[1] = "-keystore"
	args[2] = keyFile
	args[3] = "-storepass"
	args[4] = password
	args[5] = "-digestalg"
	args[6] = "SHA1"
	args[7] = "-sigalg"
	args[8] = "MD5withRSA"
	args[9] = "-signedjar"
	args[10] = signApk
	args[11] = unsignApk
	args[12] = aliasName
	cmd := exec.Command(signTool, args...)

	//fmt.Printf("%v", args)
	//var buf []byte
	//buf, err = cmd.Output()
	_, err = cmd.Output()

	//if err != nil {
	//	fmt.Fprintf(os.Stderr, "The command failed to perform: %s", err)
	//}
	//fmt.Fprintf(os.Stdout, "Result: %s", buf)
	//close(c)
	return
}

func test() {
	UnCompileApk("uncompile.apk", "uncompile")
	CompileApk("uncompile", "compile.apk")
	//c = make(chan bool)
	go ApkSign("hygame.keystore", "hygame147", "unsign.apk", "sign.apk", "hygame")
	//<-c
	//fmt.Println("Done!")
}