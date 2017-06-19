/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
	"github.com/golang/glog"
	"golang.org/x/net/context"
)

type etcdFlags struct {
	etcdAddress *string
	certFile    *string
	keyFile     *string
	caFile      *string
	dryRun      *bool
}

func generateV2ClientConfig(flags *etcdFlags) (client.Client, error) {
	if *(flags.etcdAddress) == "" {
		return nil, fmt.Errorf("--etcd-address flag is required")
	}

	tls := transport.TLSInfo{
		CAFile:   *(flags.caFile),
		CertFile: *(flags.certFile),
		KeyFile:  *(flags.keyFile),
	}

	tr, err := transport.NewTransport(tls, 30*time.Second)
	if err != nil {
		return nil, err
	}

	cfg := client.Config{
		Transport:               tr,
		Endpoints:               []string{*(flags.etcdAddress)},
		HeaderTimeoutPerRequest: 30 * time.Second,
	}

	return client.New(cfg)
}

func generateV3ClientConfig(flags *etcdFlags) (*clientv3.Config, error) {
	if *(flags.etcdAddress) == "" {
		return nil, fmt.Errorf("--etcd-address flag is required")
	}

	c := &clientv3.Config{
		Endpoints: []string{*(flags.etcdAddress)},
	}

	var cfgtls *transport.TLSInfo
	tlsinfo := transport.TLSInfo{}
	if *(flags.certFile) != "" {
		tlsinfo.CertFile = *(flags.certFile)
		cfgtls = &tlsinfo
	}

	if *(flags.keyFile) != "" {
		tlsinfo.KeyFile = *(flags.keyFile)
		cfgtls = &tlsinfo
	}

	if *(flags.caFile) != "" {
		tlsinfo.CAFile = *(flags.caFile)
		cfgtls = &tlsinfo
	}

	if cfgtls != nil {
		clientTLS, err := cfgtls.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("Error while creating etcd client: %v", err)
		}
		c.TLS = clientTLS
	}
	return c, nil
}

func getLeafKeys(nodes client.Nodes) []string {
	leaves := make([]string, 0, 0)
	for _, node := range nodes {
		if node.Dir {
			leaves = append(leaves, getLeafKeys(node.Nodes)...)
			continue
		}
		leaves = append(leaves, node.Key)
	}
	return leaves
}

func getV2Keys(keysAPI client.KeysAPI) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	lsresp, err := keysAPI.Get(ctx, "", &client.GetOptions{Sort: true, Quorum: true, Recursive: true})
	cancel()
	if err != nil {
		return nil, err
	}

	return getLeafKeys(lsresp.Node.Nodes), nil
}

func getV3Keys(c *clientv3.Client) ([]string, error) {
	// get "" --from-key --keys-only
	resp, err := c.KV.Get(context.Background(), "/", clientv3.WithFromKey())
	if err != nil {
		return nil, err
	}
	var keys []string
	for _, item := range resp.Kvs {
		keys = append(keys, string(item.Key))
	}
	return keys, nil
}

func deleteV3Key(c *clientv3.Client, key string) error {
	opts := []clientv3.OpOption{}
	_, err := c.KV.Delete(context.Background(), key, opts...)
	if err != nil {
		return err
	}
	return nil
}

func getDeletedv3Keys(v2keys, v3keys []string) []string {
	var deleted []string
	for _, v3key := range v3keys {
		found := false
		for _, v2key := range v2keys {
			if v2key == v3key {
				found = true
				break
			}
		}
		if !found {
			deleted = append(deleted, v3key)
		}
	}
	return deleted
}

func main() {
	flags := &etcdFlags{
		etcdAddress: flag.String("etcd-address", "", "Etcd address"),
		certFile:    flag.String("cert", "", "identify secure client using this TLS certificate file"),
		keyFile:     flag.String("key", "", "identify secure client using this TLS key file"),
		caFile:      flag.String("cacert", "", "verify certificates of TLS-enabled secure servers using this CA bundle"),
		dryRun:      flag.Bool("dry-run", false, "Just display v3 keys that would got removed"),
	}

	flag.Parse()

	v2ClientConfig, err := generateV2ClientConfig(flags)
	if err != nil {
		glog.Fatal(err)
	}

	keysAPI := client.NewKeysAPI(v2ClientConfig)

	v3ClientConfig, err := generateV3ClientConfig(flags)
	if err != nil {
		glog.Fatal(err)
	}

	v3Client, err := clientv3.New(*v3ClientConfig)
	if err != nil {
		glog.Fatal(err)
	}

	// Get all keys
	keys, err := getV2Keys(keysAPI)
	if err != nil {
		glog.Fatal(err)
	}

	v3keys, err := getV3Keys(v3Client)
	if err != nil {
		glog.Fatal(err)
	}

	deleted := getDeletedv3Keys(keys, v3keys)

	fmt.Printf("# v2 keys in total: %v\n", len(keys))
	fmt.Printf("# v3 keys in total: %v\n", len(v3keys))
	fmt.Printf("# v3 keys not present in v2 (total: %v):\n\n", len(deleted))

	if *(flags.dryRun) {
		fmt.Printf("# List of keys that would be deleted:\n")
	}

	for _, key := range deleted {
		if *(flags.dryRun) {
			fmt.Println(key)
		} else {
			if err := deleteV3Key(v3Client, key); err != nil {
				fmt.Printf("Unable to delete key %v: %v\n", key, err)
				continue
			}
			fmt.Printf("%q key deleted\n", key)
		}
	}
}
