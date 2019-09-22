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

package SpringCloudStarterConsul

import (
	"fmt"
	"net/http"
	"github.com/satori/go.uuid"
	"github.com/didi/go-spring/spring-core"
	ConsulApi "github.com/hashicorp/consul/api"
	"github.com/go-spring/go-spring-boot/spring-boot"
	"github.com/go-spring/go-spring-cloud/spring-cloud-consul"
)

func init() {
	SpringBoot.RegisterModule(func(ctx SpringCore.SpringContext) {
		ctx.RegisterBean(new(SpringCloudConsul.ConsulDiscoveryConfig))
		ctx.RegisterBean(new(SpringCloudConsul.ConsulDiscoveryClient))
	})
}

type ConsulDiscoveryClientWrapper struct {
	*SpringCloudConsul.ConsulDiscoveryClient
}

func (client *ConsulDiscoveryClientWrapper) OnStartApplication(context SpringBoot.SpringApplicationContext) {
	go client.RegisterServer(context)
}

func (client *ConsulDiscoveryClientWrapper) OnStopApplication(context SpringBoot.SpringApplicationContext) {

}

func (client *ConsulDiscoveryClientWrapper) RegisterServer(context SpringBoot.SpringApplicationContext) error {

	registration := new(ConsulApi.AgentServiceRegistration)
	registration.ID = uuid.NewV4().String()
	//registration.Name = context.AppName
	//registration.Address = context.ServerIP
	//registration.Port = int(context.ServerPort)

	checkAddress := fmt.Sprintf("%s:%d", registration.Address, client.Config.CheckPort)
	checkUrl := fmt.Sprintf("http://%s%s", checkAddress, client.CheckPath)

	registration.Check = &ConsulApi.AgentServiceCheck{
		HTTP:                           checkUrl,
		Timeout:                        "1s",
		Interval:                       "1s",
		DeregisterCriticalServiceAfter: "10s", //check失败后删除本服务
	}

	err := client.Client.Agent().ServiceRegister(registration)
	if err != nil {
		return err
	}

	if client.CheckHandler == nil {
		http.HandleFunc(client.CheckPath, func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "ok")
		})
	}

	// TODO 如何实现优雅的退出？
	return http.ListenAndServe(checkAddress, nil)
}
