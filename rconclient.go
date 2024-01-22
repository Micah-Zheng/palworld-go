package main

import (
	"fmt"
	"log"

	"github.com/gorcon/rcon"
)

// RconClient 结构体，用于存储RCON连接和配置信息
type RconClient struct {
	Conn       *rcon.Conn
	BackupTask *BackupTask
}

// NewRconClient 创建一个新的RCON客户端
func NewRconClient(address, password string, BackupTask *BackupTask) *RconClient {
	conn, err := rcon.Dial(address, password)
	if err != nil {
		log.Fatalf("无法连接到RCON服务器: %v", err)
	}
	return &RconClient{
		Conn:       conn,
		BackupTask: BackupTask,
	}
}

// Close 关闭RCON连接
func (client *RconClient) Close() {
	err := client.Conn.Close()
	if err != nil {
		log.Printf("关闭RCON连接时发生错误: %v", err)
	}
}

// 重启服务器
func RestartServer(RconClient *RconClient) error {
	if _, err := RconClient.Conn.Execute("broadcast Auto_Reboot_Initialized"); err != nil {
		return fmt.Errorf("error broadcasting restart initialization: %w", err)
	}
	if _, err := RconClient.Conn.Execute("save"); err != nil {
		return fmt.Errorf("error saving game state: %w", err)
	}
	if _, err := RconClient.Conn.Execute("shutdown 300 Server_is_going_to_reboot_in_5_minutes"); err != nil {
		return fmt.Errorf("error executing shutdown: %w", err)
	}
	return nil
}

func HandleMemoryUsage(threshold float64, RconClient *RconClient) {
	// 广播内存超阈值的警告
	if _, err := RconClient.Conn.Execute(fmt.Sprintf("broadcast Memory_Is_Above_%v%%", threshold)); err != nil {
		log.Printf("Error broadcasting memory threshold alert: %v", err)
	}

	// 保存游戏状态
	if _, err := RconClient.Conn.Execute("save"); err != nil {
		log.Printf("Error saving game state: %v", err)
	}

	// 安排服务器重启
	if _, err := RconClient.Conn.Execute("shutdown 60 Reboot_In_60_Seconds"); err != nil {
		log.Printf("Error executing shutdown: %v", err)
	}

	RconClient.BackupTask.RunBackup()
}
