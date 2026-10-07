package coordinator

import (
	"context"
	"fmt"
	"snowflake/pkg/snowflake"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	prefix = "/snowflake/workers/"
)

func AllocateWorker(client *clientv3.Client, workerIdentity string, ttl int64) (int, *clientv3.LeaseID, error) {
	ctx := context.Background()

	leaseResp, err := client.Grant(ctx, ttl)
	if err != nil {
		return -1, nil, err
	}
	leaseID := leaseResp.ID

	maxWorkerID := int(snowflake.WorkerIDMask)
	for id := 0; id < maxWorkerID; id++ {
		key := fmt.Sprintf("%s%d", prefix, id)

		txn := client.Txn(ctx).
			If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
			Then(clientv3.OpPut(key, workerIdentity, clientv3.WithLease(leaseID))).
			Else()

		resp, err := txn.Commit()
		if err != nil {
			client.Revoke(ctx, leaseID)
			return -1, nil, err
		}
		if resp.Succeeded {
			keepAlive, err := client.KeepAlive(ctx, leaseID)
			if err != nil {
				client.Revoke(ctx, leaseID)
				return -1, nil, err
			}
			go func() {
				for range keepAlive {
					//this boy is hungryyyy
				}
			}()
			return id, &leaseID, nil
		}
	}
	client.Revoke(ctx, leaseID)
	return -1, nil, fmt.Errorf("no free worker IDs left (0-%d)", maxWorkerID)
}
