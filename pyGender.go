package main

import (
	"encoding/base64"
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorilla/websocket"
	"io"
	"io/ioutil"
	"os"
	"pyGender/controllers"
	"pyGender/models"
	"time"
)

const (
	STATUS_FIRST_FRAME    = 0
	STATUS_CONTINUE_FRAME = 1
	STATUS_LAST_FRAME     = 2
	HttpCodeSuccessHandshake = 101  //握手成功返回的httpcode
)
 const (
 	configpath = "conf"
	configname = "pygender.yaml"
 )


func main() {

	//conf
	models.GlobalConfig = viper.New()
	models.GlobalConfig.AddConfigPath(configpath)     //设置读取的文件路径
	models.GlobalConfig.SetConfigName(configname) //设置读取的文件名
	models.GlobalConfig.SetConfigType("yaml")   //设置文件的类型
	err := models.GlobalConfig.ReadInConfig()
	if err != nil {
		fmt.Println(fmt.Errorf("Fatal error when reading %s config file:%s", configname, err))
		os.Exit(1)
	}

	models.InitLogger()


	hostUrl := models.GlobalConfig.GetString("Server.hostUrl")

	apiKey := models.GlobalConfig.GetString("Server.apiKey")
	fmt.Println("Server.apiKey:", apiKey)

	apiSecret := models.GlobalConfig.GetString("Server.apiSecret")
	fmt.Println("Server.apiSecret:", apiSecret)

	appid := models.GlobalConfig.GetString("Server.appid")
	fmt.Println("Server.appid:", appid)

	filename := models.GlobalConfig.GetString("AudioFile.name")
	fmt.Println("AudioFile.name:", filename)

	filepath := models.GlobalConfig.GetString("AudioFile.path")
	fmt.Println("AudioFile.path:", filepath)
	var file string
	file += filepath+filename

	d := websocket.Dialer{
		HandshakeTimeout: 1 * time.Second,
	}
	//握手并建立websocket 连接
	conn, resp, err := d.Dial(controllers.AssembleAuthUrl(hostUrl, apiKey, apiSecret), nil)  //导出的函数首字母必须大写
	if err != nil {
		panic(err)
		if resp.StatusCode != HttpCodeSuccessHandshake {
			b, _ := ioutil.ReadAll(resp.Body)
			models.Logger.Error("handshake failed:",zap.String("ErrInfo:",string(b)),zap.String("StatusCode:",fmt.Sprintf("%s",resp.StatusCode)))
		//	fmt.Printf("handshake failed:message=%s,httpCode=%d\n", string(b), resp.StatusCode)
		}
		return
	}
	//打开音频文件
	audioFile, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	//开启协程，发送数据
	go func() {
		var frameSize = models.GlobalConfig.GetInt("AudioFile.framesize")                 //每一帧的音频大小
		var intervel = time.Duration(models.GlobalConfig.GetInt("AudioFile.intervel"))  * time.Millisecond //发送音频间隔
		var status = STATUS_FIRST_FRAME      //音频的状态信息，标识音频是第一帧，还是中间帧、最后一帧
		var buffer = make([]byte, frameSize)
		for {
			len, err := audioFile.Read(buffer)
			if err != nil {
				if err == io.EOF { //文件读取完了，改变status = STATUS_LAST_FRAME
					status = STATUS_LAST_FRAME
				} else {
					panic(err)
				}
			}
			switch status {
			case STATUS_FIRST_FRAME: //发送第一帧音频，带business 参数
				frameData := map[string]interface{}{
					"common": map[string]interface{}{
						"app_id": appid, //appid 必须带上，只需第一帧发送
					},
					"business": map[string]interface{}{ //business 参数，只需一帧发送
						"rate": 16000,
						"aue": "raw",
					},
					"data": map[string]interface{}{
						"status": STATUS_FIRST_FRAME, //第一帧音频status要为 0
						"audio":  base64.StdEncoding.EncodeToString(buffer[:len]),
					},
				}
				conn.WriteJSON(frameData)
				status = STATUS_CONTINUE_FRAME
			case STATUS_CONTINUE_FRAME:
				frameData := map[string]interface{}{
					"data": map[string]interface{}{
						"status": STATUS_CONTINUE_FRAME, // 中间音频status 要为1
						"audio":  base64.StdEncoding.EncodeToString(buffer[:len]),
					},
				}
				conn.WriteJSON(frameData)
			case STATUS_LAST_FRAME:
				frameData := map[string]interface{}{
					"data": map[string]interface{}{
						"status": STATUS_LAST_FRAME, // 最后一帧音频status 一定要为2 且一定发送
						"audio":  base64.StdEncoding.EncodeToString(buffer[:len]),
					},
				}
				conn.WriteJSON(frameData)
				goto end
			}
			//模拟音频采样间隔
			time.Sleep(intervel)
		}
	end:
	}()
	//获取返回的数据
	for {
		var resp = &models.RespData{}
		err := conn.ReadJSON(resp)
		if err != nil {
			models.Logger.Error("read message error",zap.String("err : ",err.Error()))
			//fmt.Println("read message error:", err)
			break
		}
		//fmt.Println(resp)
		models.Logger.Info("read message success.")

		if resp.Code == 0 {
			if resp.Data != nil {
				if result:=resp.Data.Result ;result!= nil {
					models.Logger.Info("Judging results",zap.String("result",fmt.Sprintf("%s", result)))
					//fmt.Printf("result is :%+v \n",result)
					// todo
				}
				if resp.Data.Status == 2 { //当返回的数据status=2时，表示数据已经全部返回，这时候应该结束本次会话
					break
				}
			}
		} else {
			models.Logger.Error("Error",zap.Int("err code : ",resp.Code),zap.String("message: ",resp.Message))
			//fmt.Println("Error:",resp.Code, "|", resp.Message)
		}
	}
	conn.Close()
}

