package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/hashicorp/serf/serf"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

func main() {
	// cache, err := bigcache.New(context.Background(), bigcache.DefaultConfig(10*time.Minute))
	// if err != nil {
	// 	log.Fatalf("bigcache.New err : %v", err)
	// }

	// d := NewDataBase(cache)

	// d.SetValue("", []byte{})
	// d.GetValue("")
	fmt.Println(os.Getenv("ADDR"))
	cluster, err := setupCluster(
		os.Getenv("ADDR"),
		os.Getenv("CLUSTER_ADDR"),
		os.Getenv("PORT"),
		os.Getenv("CLUSTER_PORT"))

	if err != nil {
		log.Fatalf("setupCluster err : %v", err)
	}

	defer cluster.Leave()

	// theOneAndOnlyNumber := InitTheNumber(42)
	// LaunchHTTPAPI(theOneAndOnlyNumber)

	// ctx := context.Background()
	// if name, err := os.Hostname(); err == nil {
	// 	ctx = context.WithValue(ctx, "name", name)
	// }

	// debugDataPrinterTicker := time.Tick(time.Second * 5)
	numberBroadcastTicker := time.Tick(time.Second * 2)
	for {
		select {
		case <-numberBroadcastTicker:
			fmt.Println("check memeber")
			_ = getOtherMembers(cluster)

			// ctx, _ := context.WithTimeout(ctx, time.Second*2)
			// go notifyOthers(ctx, members, theOneAndOnlyNumber)
			// case <-debugDataPrinterTicker:
			// 	log.Printf("Members: %v\n", cluster.Members())

			// 	// curVal, curGen := theOneAndOnlyNumber.getValue()
			// 	log.Printf("State: Val: %v Gen: %v\n", curVal, curGen)
		}
	}
}

func setupCluster(addr, clusterAddr, port, clusterPort string) (*serf.Serf, error) {
	conf := serf.DefaultConfig()
	conf.Init()
	fmt.Println(addr, clusterAddr, port, clusterPort)
	conf.MemberlistConfig.AdvertiseAddr = addr
	conf.MemberlistConfig.BindPort, _ = strconv.Atoi(port)
	conf.MemberlistConfig.AdvertisePort, _ = strconv.Atoi(port)
	cluster, err := serf.Create(conf)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't create cluster")
	}

	_, err = cluster.Join([]string{clusterAddr + ":" + port}, true)
	if err != nil {
		log.Printf("Couldn't join cluster, starting own: %v\n", err)
	}

	return cluster, nil
}

func LaunchHTTPAPI(db *oneAndOnlyNumber) {
	go func() {
		m := mux.NewRouter()
		m.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
			val, _ := db.getValue()
			fmt.Fprintf(w, "%v", val)
		})

		m.HandleFunc("/set/{value}", func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			newVal, err := strconv.Atoi(vars["newVal"])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "%v", err)
				return
			}

			db.setValue(newVal)

			fmt.Fprintf(w, "%v", newVal)
		})

		m.HandleFunc("/notify/{curVal}/{curGeneration}", func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			curVal, err := strconv.Atoi(vars["curVal"])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "%v", err)
				return
			}
			curGeneration, err := strconv.Atoi(vars["curGeneration"])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "%v", err)
				return
			}

			if changed := db.notifyValue(curVal, curGeneration); changed {
				log.Printf(
					"NewVal: %v Gen: %v Notifier: %v",
					curVal,
					curGeneration,
					r.URL.Query().Get("notifier"))
			}
			w.WriteHeader(http.StatusOK)
		})
		log.Fatal(http.ListenAndServe(":8080", m))
	}()
}

func getOtherMembers(cluster *serf.Serf) []serf.Member {
	members := cluster.Members()
	fmt.Println(len(members))
	fmt.Println(cluster.LocalMember().Addr, cluster.LocalMember().Name)
	for i := 0; i < len(members); {
		fmt.Println(members[i].Name)

		if members[i].Name == cluster.LocalMember().Name || members[i].Status != serf.StatusAlive {
			if i < len(members)-1 {
				members = append(members[:i], members[i+1:]...)
			} else {
				members = members[:i]
			}
		} else {
			i++
		}
	}
	return members
}

func notifyOthers(ctx context.Context, otherMembers []serf.Member, db *oneAndOnlyNumber) {
	g, ctx := errgroup.WithContext(ctx)

	if len(otherMembers) <= MembersToNotify {
		for _, member := range otherMembers {
			curMember := member
			g.Go(func() error {
				return notifyMember(ctx, curMember.Addr.String(), db)
			})
		}
	} else {
		number := rand.Int()
		randIndex := number % len(otherMembers)
		for i := 0; i < MembersToNotify; i++ {
			g.Go(func() error {
				return notifyMember(
					ctx,
					otherMembers[(randIndex+i)%len(otherMembers)].Addr.String(),
					db)
			})
		}
	}

	err := g.Wait()
	if err != nil {
		log.Printf("Error when notifying other members: %v", err)
	}
}

func notifyMember(ctx context.Context, addr string, db *oneAndOnlyNumber) error {
	val, gen := db.getValue()
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%v:8080/notify/%v/%v?notifier=%v", addr, val, gen, ctx.Value("name")), nil)
	if err != nil {
		return errors.Wrap(err, "Couldn't create request")
	}
	req = req.WithContext(ctx)

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "Couldn't make request")
	}
	return nil
}

const MembersToNotify = 2

type oneAndOnlyNumber struct {
	num        int
	generation int
	numMutex   sync.RWMutex
}

func InitTheNumber(val int) *oneAndOnlyNumber {
	return &oneAndOnlyNumber{
		num: val,
	}
}

func (n *oneAndOnlyNumber) setValue(newVal int) {
	n.numMutex.Lock()
	defer n.numMutex.Unlock()
	n.num = newVal
	n.generation = n.generation + 1
}

func (n *oneAndOnlyNumber) getValue() (int, d int) {
	n.numMutex.RLock()
	defer n.numMutex.RUnlock()
	return n.num, n.generation
}

func (n *oneAndOnlyNumber) notifyValue(curVal int, curGeneration int) bool {
	if curGeneration > n.generation {
		n.numMutex.Lock()
		defer n.numMutex.Unlock()
		n.generation = curGeneration
		n.num = curVal
		return true
	}
	return false
}
