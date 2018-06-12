package main

import(
    "test/keyValStore"
    "flag"
    "os"
    "os/signal"
    "test/utils"
    "fmt"
)

// Arguments:
//   - i: replica number
//   - config: path name of config
func main() {
    configPathPtr := flag.String("config", "config", "Path to config file")
    iPtr := flag.Int("i", 0, "Replica number, indexed by line in config file")

    flag.Parse()

    configPath := *configPathPtr
    i := *iPtr

    config, err := utils.ReadConfig(configPath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error reading config at %s: %s", config, err)
        return
    }

    servers := config.Servers
    keyValStore.StartNode(keyValStore.CreateWorkers(1)[0], servers, i)

    // Wait for CTRL-C
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    <-c
}
