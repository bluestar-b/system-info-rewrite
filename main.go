package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func getSystemInfo() map[string]interface{} {
	cpuInfo, _ := cpu.Info()
	cpuPercent, _ := cpu.Percent(1, true)
	memoryInfo, _ := mem.VirtualMemory()
	diskUsage, _ := disk.Usage("/")
	netIOCounters, _ := net.IOCounters(false)

	systemInfo := map[string]interface{}{
		"cpu_count":   len(cpuInfo),
		"cpu_percent": cpuPercent,
		"memory_info": map[string]interface{}{
			"total":     memoryInfo.Total,
			"available": memoryInfo.Available,
			"percent":   memoryInfo.UsedPercent,
		},
		"disk_usage": map[string]interface{}{
			"total":   diskUsage.Total,
			"used":    diskUsage.Used,
			"free":    diskUsage.Free,
			"percent": diskUsage.UsedPercent,
		},
		"net_io_counters": netIOCounters[0],
	}

	return systemInfo
}

func sendDataToClients(conn *websocket.Conn) {
	for {
		systemInfo := getSystemInfo()
		err := conn.WriteJSON(systemInfo)
		if err != nil {
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func main() {
	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "We are star stuff which has taken its destiny into its own hands.",
		})
	})

	router.GET("/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, getSystemInfo())
	})

	router.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade to WebSocket"})
			return
		}
		defer conn.Close()

		sendDataToClients(conn)
	})

	router.Run(":81")
}
