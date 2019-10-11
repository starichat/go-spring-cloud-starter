/*
 * Copyright 2012-2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package SpringCloudStarterZookeeper

import (
	"encoding/json"
	"strings"

	"github.com/go-spring/go-spring-boot/spring-boot"
	"github.com/go-spring/go-spring-cloud/spring-cloud-discovery"
	"github.com/go-spring/go-spring-cloud/spring-cloud-zookeeper"
	"github.com/go-spring/go-spring/spring-core"
	"github.com/go-spring/go-spring/spring-utils"
	"github.com/samuel/go-zookeeper/zk"
)

func init() {
	SpringBoot.RegisterModule(func(ctx SpringCore.SpringContext) {
		ctx.RegisterBean(new(SpringCloudZookeeper.ZookeeperDiscoveryConfig))

		c := new(SpringCloudZookeeper.ZookeeperDiscoveryClient)
		ctx.RegisterBean(c)

		cs := &ZookeeperDiscoveryClientWraper{c}
		ctx.RegisterBean(cs)
	})
}

type ZookeeperDiscoveryClientWraper struct {
	*SpringCloudZookeeper.ZookeeperDiscoveryClient
}

func (client *ZookeeperDiscoveryClientWraper) OnStartApplication(context SpringBoot.ApplicationContext) {

	servers := strings.Split(client.Config.Address, ",")
	conn, session, err := zk.Connect(servers, 5e9)
	if err != nil {
		//Logger.Fatalf("Can't connect: %v", err)
		panic(err)
	}

	//defer zk.Close()

	//conn.SetLogger(Logger.Logger)

	for {
		event := <-session
		if event.State == zk.StateConnected {
			break
		}
	}

	client.Conn = conn

	if err := client.RegisterServer(context); err != nil {
		panic(err)
	}
}

func (client *ZookeeperDiscoveryClientWraper) OnStopApplication(context SpringBoot.ApplicationContext) {

}

func (client *ZookeeperDiscoveryClientWraper) RegisterServer(context SpringCore.SpringContext) (err error) {

	if err = client.CreateNode(SpringCloudZookeeper.ZOOKEEPER_DISCOVERY_ROOT, nil, 0); err != nil {
		return err
	}

	path0 := SpringCloudZookeeper.ZOOKEEPER_DISCOVERY_ROOT + "/" + client.Config.AppName
	path0 = strings.Replace(path0, "-", "_", -1)

	if err = client.CreateNode(path0, nil, 0); err != nil {
		return err
	}

	var instance SpringCloudDiscovery.ServiceInstance

	instance.InstanceId = client.AppId
	instance.ServiceId = client.Config.AppName
	instance.Host = SpringUtils.LocalIPv4()
	// TODO
	//instance.Port = int(ServerPort)
	//instance.SecurePort = int(ServerPort)

	bytes, err := json.Marshal(&instance)
	if err != nil {
		return err
	}

	path1 := path0 + "/" + client.AppId
	path1 = strings.Replace(path1, "-", "_", -1)

	return client.CreateNode(path1, bytes, zk.FlagEphemeral)
}
