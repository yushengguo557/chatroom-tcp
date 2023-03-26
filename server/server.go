package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/rs/xid"
)

var (
	enteringChannel = make(chan *User)      // 进入
	leavingChannel  = make(chan *User)      // 退出
	messageChannel  = make(chan Message, 8) // 全局消息
)

// User User结构体
type User struct {
	ID             string      // 用户唯一标识 GenUserID函数生成
	Addr           string      // 用户ID地址和端口
	EnterAt        time.Time   // 用户进入时间
	MessageChannel chan string // 当前用户发送消息的通道
}

// Message Message结构体
type Message struct {
	OwnerID string // 消息的发出者ID
	Content string // 消息内容
}

func main() {
	listener, err := net.Listen("tcp", ":2023")
	log.Println("Listening on port 2023...")
	if err != nil {
		log.Println(err)
	}

	go broadcaster()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close() // 关闭连接

	// 1. 实例化User结构体创建用户user
	user := &User{
		ID:             GenUserID(),
		Addr:           conn.RemoteAddr().String(),
		EnterAt:        time.Now(),
		MessageChannel: make(chan string, 8),
	}
	log.Println("--- " + user.Addr + " has enter the chatroom." + " ---")
	// 2. 给 user 发消息
	go sendMessage(conn, user.MessageChannel)

	// 3. 给当前用户发送Welcome消息 给其他用户发送has enter消息
	user.MessageChannel <- "Welcome, " + user.ID
	messageChannel <- Message{
		OwnerID: user.ID,
		Content: "User: `" + user.ID + "` has enter.",
	}

	// 4. 当前用户写入全局用户列表
	enteringChannel <- user

	// 5. 读取用户输入并广播发送
	input := bufio.NewScanner(conn)
	for input.Scan() {
		messageChannel <- Message{
			OwnerID: user.ID,
			Content: user.ID + ": " + input.Text(),
		}
	}
	if err := input.Err(); err != nil {
		log.Println(err)
	}

	// 6. 用户离开 进入leavingChannel 并向其他所有用户发送left消息
	leavingChannel <- user
	messageChannel <- Message{
		OwnerID: user.ID,
		Content: "User: `" + user.ID + "` has left.",
	}
}

// sendMessage 给客户端发送消息
func sendMessage(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

// broadcaster 消息广播
func broadcaster() {
	users := make(map[*User]struct{})

	for {
		select {
		case user := <-enteringChannel:
			users[user] = struct{}{}
		case user := <-leavingChannel:
			delete(users, user)
			close(user.MessageChannel) // 避免 goroutine 泄露
		case msg := <-messageChannel:
			for user := range users {
				if user.ID != msg.OwnerID {
					user.MessageChannel <- msg.Content
				}
			}
		}
	}
}

func GenUserID() string {
	// 生成全局唯一的 ID
	id := xid.New()
	return id.String()
}
