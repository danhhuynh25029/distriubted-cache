package main

import (
	"context"
	"log"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/hashicorp/serf/serf"
	"github.com/pkg/errors"
)

func main() {
	cache, err := bigcache.New(context.Background(), bigcache.DefaultConfig(10*time.Minute))
	if err != nil {
		log.Fatalf("bigcache.New err : %v", err)
	}

	d := NewDataBase(cache)
	d.SetValue("", []byte{})
	d.GetValue("")

}

func setupCluster(addr, clusterAddr string) (*serf.Serf, error) {
	conf := serf.DefaultConfig()
	conf.Init()
	conf.MemberlistConfig.AdvertiseAddr = addr

	cluster, err := serf.Create(conf)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't create cluster")
	}

	_, err = cluster.Join([]string{clusterAddr}, true)
	if err != nil {
		log.Printf("Couldn't join cluster, starting own: %v\n", err)
	}

	return cluster, nil
}
