# Configuration

The majority of the configuration is handled through an Ansible inventory
directory. A sample inventory can be found at
`openshift-ansible/playbooks/openstack/sample-inventory/`.

`inventory/group_vars/all.yml` is used for OpenStack configuration,
while `inventory/group_vars/OSEv3.yml` is used for OpenShift
configuration.

Environment variables may also be used.

* [OpenStack Configuration](#openstack-configuration)
* [OpenShift Configuration](#openshift-configuration)
* [Stack Name Configuration](#stack-name-configuration)
* [DNS Configuration](#dns-configuration)
* [Kuryr Networking Configuration](#kuryr-networking-configuration)
* [Provider Network Configuration](#provider-network-configuration)
* [Multi-Master Configuration](#multi-master-configuration)
* [Provider Network Configuration](#provider-network-configuration)
* [OpenStack Credential Configuration](#openstack-credential-configuration)
* [Cinder-Backed Persistent Volumes Configuration](#cinder-backed-persistent-volumes-configuration)
* [Cinder-Backed Registry Configuration](#cinder-backed-registry-configuration)


## OpenStack Configuration

In `inventory/group_vars/all.yml`:

* `openshift_openstack_keypair_name` OpenStack keypair to use.
* Role Node Counts
  * `openshift_openstack_num_masters` Number of master nodes to create.
  * `openshift_openstack_num_infra` Number of infra nodes to create.
  * `openshift_openstack_num_nodes` Number of app nodes to create.
* Role Images
  * `openshift_openstack_default_image_name` OpenStack image used by all VMs, unless a particular role image name is specified.
  * `openshift_openstack_master_image_name`
  * `openshift_openstack_infra_image_name`
  * `openshift_openstack_cns_image_name`
  * `openshift_openstack_node_image_name`
  * `openshift_openstack_lb_image_name`
  * `openshift_openstack_etcd_image_name`
* Role Flavors
  * `openshift_openstack_default_flavor` OpenStack flavor used by all VMs, unless a particular role flavor name is specified.
  * `openshift_openstack_master_flavor`
  * `openshift_openstack_infra_flavor`
  * `openshift_openstack_cns_flavor`
  * `openshift_openstack_node_flavor`
  * `openshift_openstack_lb_flavor`
  * `openshift_openstack_etcd_flavor`
* Role Hostnames: used for customizing public names of Nova servers provisioned with a given role.
  * `openshift_openstack_master_hostname` Defaults to `master`.
  * `openshift_openstack_infra_hostname` Defaults to `infra-node`.
  * `openshift_openstack_cns_hostname` Defaults to `cns`.
  * `openshift_openstack_node_hostname` Defaults to `app-node`.
  * `openshift_openstack_lb_hostname` Defaults to `lb`.
  * `openshift_openstack_etcd_hostname` Defaults to `etcd`.
* `openshift_openstack_external_network_name` OpenStack network providing external connectivity.
* `openshift_openstack_cluster_node_labels` Custom labels for openshift cluster node groups; currently supports app and infra node groups.
The default value of this variable sets `region: primary` to app nodes and `region: infra` to infra nodes. An example of setting a customized label:

```
openshift_openstack_cluster_node_labels:
  app:
    mylabel: myvalue
```

* `openshift_openstack_provision_user_commands` Allows users to execute shell commands via cloud-init for all of the created Nova servers in the Heat stack, before they are available for SSH connections. Note that you should use [custom Ansible playbooks](./post-install.md#run-custom-post-provision-actions) whenever possible. User specified shell commands for cloud-init need to be either strings or lists:

```
- openshift_openstack_provision_user_commands:
  - set -vx
  - systemctl stop sshd # fences off ansible playbooks as we want to reboot later
  - ['echo', 'foo', '>', '/tmp/foo']
  - [ ls, /tmp/foo, '||', true ]
  - reboot # unfences ansible playbooks to continue after reboot
```

* `openshift_openstack_nodes_to_remove` The numerical indexes of app nodes that should be removed; for example, `['0', '2']`,
* Role Docker Volume Size
  * `openshift_openstack_docker_volume_size` Default Docker volume size used by all VMs, unless a particular role Docker volume size is specified. If `openshift_openstack_ephemeral_volumes` is set to `true`, the `*_volume_size` variables will be ignored and the deployment will not create any cinder volumes.
  * `openshift_openstack_docker_master_volume_size`
  * `openshift_openstack_docker_infra_volume_size`
  * `openshift_openstack_docker_cns_volume_size`
  * `openshift_openstack_docker_node_volume_size`
  * `openshift_openstack_docker_etcd_volume_size`
  * `openshift_openstack_docker_lb_volume_size`
* `openshift_openstack_flat_secgrp` Set to True if you experience issues with sec group rules quotas. It trades security for number of rules, by sharing the same set of firewall rules for master, node, etcd and infra nodes.
* `openshift_openstack_required_packages` List of additional prerequisite packages to be installed before deploying an OpenShift cluster.


## OpenShift Configuration

In `inventory/group_vars/OSEv3.yml`:

* `openshift_disable_check` List of checks to disable.
* `openshift_master_cluster_public_hostname` Custom entrypoint; for example, `api.openshift.example.com`. Note than an empty hostname does not work, so if your domain is `openshift.example.com` you cannot set this value to simply `openshift.example.com`.
* `openshift_deployment_type` Version of OpenShift to deploy; for example, `origin` or `openshift-enterprise`
* `openshift_master_default_subdomain`

Additional options can be found in this sample inventory:

https://github.com/openshift/openshift-ansible/blob/master/inventory/hosts.example


## Stack Name Configuration

By default the Heat stack created by OpenStack for the OpenShift cluster will be
named `openshift-cluster`. If you would like to use a different name then you
must set the `OPENSHIFT_CLUSTER` environment variable before running the playbooks:

```
$ export OPENSHIFT_CLUSTER=openshift.example.com
```

If you use a non-default stack name and run the openshift-ansible playbooks to update
your deployment, you must set `OPENSHIFT_CLUSTER` to your stack name to avoid errors.


## DNS Configuration

Note that the provisioning playbooks update the original Neutron subnet
created with the Heat stack to point to the configured DNS servers.
So the provisioned cluster nodes will start using those natively as
default nameservers. Technically, this allows the deployment of OpenShift
clusters without dnsmasq proxies.

In `inventory/group_vars/all.yml`:

* `openshift_openstack_clusterid` Defaults to `openshift`
* `openshift_openstack_public_dns_domain` Defaults to `example.com`

These two parameters together form the cluster's public DNS domain that all
the servers will be under; by default this domain will be `openshift.example.com`.

* `openshift_openstack_app_subdomain` Subdomain for workloads. Defaults to `apps`.

* `openshift_openstack_public_hostname_suffix` Empty by default.
* `openshift_openstack_private_hostname_suffix` Empty by default.

If you want to use two sets of hostnames for public and private/prefixed DNS
records for your externally managed public DNS server, you can specify the
`openshift_openstack_*_hostname_suffix` parameters. These suffixes are added to
the nsupdate records sent to the external DNS server. Note that the real hostnames,
Nova servers, and ansible hostnames and inventory variables are not be updated.
The deployment may be done on arbitrary named hosts with the hostnames managed by
cloud-init. Inventory hostnames will ignore these suffixes.

* `openshift_openstack_dns_nameservers` List of DNS servers accessible from all the created Nova servers. These will provide the internal name resolution for your OpenShift nodes (as well as upstream name resolution for installing packages, etc.).

* `openshift_use_dnsmasq` Controls whether dnsmasq is deployed or not.By default, dnsmasq is deployed and comes as the hosts' /etc/resolv.conf file first nameserver entry that points to the local host instance of the dnsmasq daemon that in turn proxies DNS requests to the authoritative DNS server. When Network Manager is enabled for provisioned cluster nodes, which is normally the case, you should not change the defaults and always deploy dnsmasq.

* `openshift_openstack_external_nsupdate_keys` Describes an external authoritative DNS server(s) processing dynamic records updates in the public only cluster view. For example:

```
    openshift_openstack_external_nsupdate_keys:
      public:
        key_secret: <some nsupdate key>
        key_algorithm: 'hmac-md5'
        key_name: 'update-key'
        server: <public DNS server IP>
```

Here, for the public view section, we specified another key algorithm and
optional `key_name`, which normally defaults to the cluster's DNS domain.
This just illustrates a compatibility mode with a DNS service deployed
by OpenShift on OSP10 reference architecture, and used in a mixed mode with
another external DNS server.


## Kuryr Networking Configuration

Kuryr is an SDN that uses OpenStack Neutron. This prevents the double overlay
overhead one would get when running OpenShift on OpenStack using the default
OpenShift SDN.

https://docs.openstack.org/kuryr-kubernetes/latest/readme.html

### OpenStack Requirements

Kuryr has a few additional requirements on the underlying OpenStack deployment:

* The Trunk Ports extension must be enabled:
  * https://docs.openstack.org/neutron/pike/admin/config-trunking.html
  * Make sure to restart `neutron-server` after you change the configuration
* Neutron must use the Open vSwitch firewall driver:
  * https://docs.openstack.org/neutron/pike/admin/config-ovsfwdriver.html
  * Make sure to restart `neutron-openvswitch-agent` after the config change
* A Load Balancer as a Service (implementing LBaaS v2 API) must be available
  * Octavia is the only supported solution right now
  * You could try the native Neutron LBaaSv2 but it is deprecated and buggy

We recommend you use the Queens or newer release of OpenStack.


### Necessary Kuryr Options

This is is the minimum you need to set (in `group_vars/all.yml`):

```yaml
openshift_use_kuryr: true
openshift_use_openshift_sdn: false
os_sdn_network_plugin_name: cni
openshift_node_proxy_mode: userspace
openshift_hosted_manage_registry: false
use_trunk_ports: true

openshift_master_open_ports:
- service: dns tcp
  port: 53/tcp
- service: dns udp
  port: 53/udp
openshift_node_open_ports:
- service: dns tcp
  port: 53/tcp
- service: dns udp
  port: 53/udp

kuryr_openstack_public_subnet_id: <public/external subnet UUID>
```

The `kuryr_openstack_public_subnet_id` value must be set to the UUID of the
public subnet in your OpenStack. In other words, the subnet with the Floating
IP range defined. It corresponds to the public network, which is often called
`public`, `external` or `ext-net`.

**NOTE**: A lot of OpenStack deployments do not make the public subnet
accessible to regular users. We will add an option to enter the network here
instead. This issue is tracked at:

https://github.com/openshift/openshift-ansible/issues/7383

### Port pooling

It is possible to pre-create Neutron ports for later use. This means that
several ports (each port will be attached to an OpenShift pod) would be created
at once. This will speed up individual pod creation at the cost of having a few
extra ports that are not currently in use.

For more information on the Kuryr port pools, check out the Kuryr
documentation:

https://docs.openstack.org/kuryr-kubernetes/latest/installation/ports-pool.html

To disable this feature, you must set:

```yaml
kuryr_openstack_enable_pools: false
```

You can control the port pooling characteristics with these options:

```yaml
kuryr_openstack_pool_max: 0
kuryr_openstack_pool_min: 1
kuryr_openstack_pool_batch: 5
kuryr_openstack_pool_update_frequency: 20
`openshift_kuryr_precreate_subports: 5`
```

Note in the last variable you specify the number of subports that will
be created per trunk port, i.e., per pool.

In order to enable the kuryr `multi-pool` driver support, we need to tag
the nodes with their corresponding `pod_vif` labels so that the right kuryr
pool driver is used for each VM/node. To do that, uncomment:

```
openshift_openstack_cluster_node_labels:
  app:
    region: primary
    pod_vif: nested-vlan
  infra:
    region: infra
    pod_vif: nested-vlan
```


### Deploying OpenShift Registry

Since we've disabled the OpenShift registry creation, you will have to create
it manually afterwards. SSH to a master node and run this as root:

```yaml
oadm registry --config=/etc/origin/master/admin.kubeconfig --service-account=registry
```

For more information (e.g. how to use a specific storage backend), please
follow the OpenShift documentation on the registry:

https://docs.openshift.org/latest/install_config/registry/index.html


## API and Router Load Balancing

A production deployment should contain more then one master and infra node and
have a load balancer in front of them.

The playbooks will not create any load balancer by default. Even if you do
request multiple masters.

You can opt into that if you want though. There are two options: a VM-based
load balancer and OpenStack's Load Balancer as a Service.

### Load Balancer as a Service

If your OpenStack supports Load Balancer as a Service (LBaaS) provided by the
Octavia project, our playbooks can set it up automatically.

Put this in your `inventory/group_vars/all.yml`:

    openshift_openstack_use_lbaas_load_balancer: true

This will create two load balancers: one for the API and UI console and the
other for the OpenShift router. Each will have its own public IP address.

### VM-based Load Balancer

If you can't use OpenStack's LBaaS, we can create and configure a virtual
machine running HAProxy to serve as one.

Put this in your `inventory/group_vars/all.yml`:

    openshift_openstack_use_vm_load_balancer: true

**WARNING** this VM will only handle the API and UI requests, *not* the
OpenShift routes.

That means, if you have more than one infra node, you will have to balance them
externally. It is not recommended to use this option in production.

### No Load Balancer

If you specify neither `openshift_openstack_use_lbaas_load_balancer` nor
`openshift_openstack_use_vm_load_balancer`, the resulting OpenShift cluster
will have no load balancing configured out of the box.

This is regardless of how many master or infra nodes you create.

In this mode, you are expected to configure and maintain a load balancer
yourself.

However, the cluster is usable without a load balancer as well. To talk to the
API or UI, connect to any of the master nodes. For the OpenShift routes, use
any of the infra nodes.

### Public Cluster Endpoints

In either of these cases (LBaaS, VM HAProxy, no LB) the public addresses to
access the cluster's API and router will be printed out at the end of the
playbook.

If you want to get them out explicitly, run the following playbook with the
same arguments (private key, inventories, etc.) as your provision/install ones:

    playbooks/openstack/inventory.py openshift-ansible/playbooks/openstack/openshift-cluster/cluster-info.yml

These addresses will depend on the load balancing solution. For LBaaS, they'll
be the the floating IPs of the load balancers. In the VM-based solution,
the API address will be the public IP of the load balancer VM and the router IP
will be the address of the first infra node that was created. If no load
balancer is selected, the API will be the address of the first master node and
the router will be the address of the first infra node.

This means that regardless of the load balancing solution, you can use these
two entries to provide access to your cluster.



## Provider Network Configuration

Normally, the playbooks create a new Neutron network and subnet and attach
floating IP addresses to each node. If you have a provider network set up, this
is all unnecessary as you can just access servers that are placed in the
provider network directly.

Note that this will not update the nodes' DNS, so running openshift-ansible
right after provisioning will fail (unless you're using an external DNS server
your provider network knows about). You must make sure your nodes are able to
resolve each other by name.

In `inventory/group_vars/all.yml`:

* `openshift_openstack_provider_network_name` Provider network name. Setting this will cause the `openshift_openstack_external_network_name` and `openshift_openstack_private_network_name` parameters to be ignored.


## OpenStack Credential Configuration

Some features require you to configure OpenStack credentials. In `inventory/group_vars/OSEv3.yml`:

* `openshift_cloudprovider_kind: openstack
* `openshift_cloudprovider_openstack_auth_url: "{{ lookup('env','OS_AUTH_URL') }}"
* `openshift_cloudprovider_openstack_username: "{{ lookup('env','OS_USERNAME') }}"
* `openshift_cloudprovider_openstack_password: "{{ lookup('env','OS_PASSWORD') }}"
* `openshift_cloudprovider_openstack_tenant_name: "{{ lookup('env','OS_PROJECT_NAME') }}"
* `openshift_cloudprovider_openstack_domain_name: "{{ lookup('env','OS_USER_DOMAIN_NAME') }}"

For more information, consult the [Configuring for OpenStack page in the OpenShift documentation][openstack-credentials].

[openstack-credentials]: https://docs.openshift.org/latest/install_config/configuring_openstack.html#install-config-configuring-openstack

**NOTE** the OpenStack integration currently requires DNS to be configured and
running and the `openshift_hostname` variable must match the Nova server name
for each node. The cluster deployment will fail without it. If you use the
provided OpenStack dynamic inventory and configure the
`openshift_openstack_dns_nameservers` Ansible variable, this will be handled
for you.


## Cinder-Backed Persistent Volumes Configuration

In addition to [setting up OpenStack credentials](#openstack-credential-configuration),
you must set the following in `inventory/group_vars/OSEv3.yml`:

* `openshift_cloudprovider_openstack_blockstorage_version`: v2

The Block Storage version must be set to `v2`, because OpenShift does not support
the v3 API yet and the version detection currently does not work.

After a successful deployment, the cluster will be configured for Cinder persistent
volumes.

### Validation

1. Log in and create a new project (with `oc login` and `oc new-project`)
2. Create a file called `cinder-claim.yaml` with the following contents:

```yaml
apiVersion: "v1"
kind: "PersistentVolumeClaim"
metadata:
  name: "claim1"
spec:
  accessModes:
    - "ReadWriteOnce"
  resources:
    requests:
      storage: "1Gi"
```
3. Run `oc create -f cinder-claim.yaml` to create the Persistent Volume Claim object in OpenShift
4. Run `oc describe pvc claim1` to verify that the claim was created and its Status is `Bound`
5. Run `openstack volume list`
   * A new volume called `kubernetes-dynamic-pvc-UUID` should be created
   * Its size should be `1`
   * It should not be attached to any server
6. Create a file called `mysql-pod.yaml` with the following contents:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: mysql
  labels:
    name: mysql
spec:
  containers:
    - resources:
        limits :
          cpu: 0.5
      image: openshift/mysql-55-centos7
      name: mysql
      env:
        - name: MYSQL_ROOT_PASSWORD
          value: yourpassword
        - name: MYSQL_USER
          value: wp_user
        - name: MYSQL_PASSWORD
          value: wp_pass
        - name: MYSQL_DATABASE
          value: wp_db
      ports:
        - containerPort: 3306
          name: mysql
      volumeMounts:
        - name: mysql-persistent-storage
          mountPath: /var/lib/mysql/data
  volumes:
    - name: mysql-persistent-storage
      persistentVolumeClaim:
        claimName: claim1
```

7. Run `oc create -f mysql-pod.yaml` to create the pod
8. Run `oc describe pod mysql`
   * Its events should show that the pod has successfully attached the volume above
   * It should show no errors
   * `openstack volume list` should show the volume attached to an OpenShift app node
   * NOTE: this can take several seconds
9. After a while, `oc get pod` should show the `mysql` pod as running
10. Run `oc delete pod mysql` to remove the pod
   * The Cinder volume should no longer be attached
11. Run `oc delete pvc claim1` to remove the volume claim
   * The Cinder volume should be deleted


## Cinder-Backed Registry Configuration

You can use a pre-existing Cinder volume for the storage of your
OpenShift registry. To do that, you need to have a Cinder volume.
You can create one by running:

```
openstack volume create --size <volume size in gb> <volume name>
```

Alternatively, the playbooks can create the volume created automatically if you
specify its name and size.

In either case, you have to [set up OpenStack
credentials](#openstack-credential-configuration), and then set the following
in `inventory/group_vars/OSEv3.yml`:

* `openshift_hosted_registry_storage_kind`: openstack
* `openshift_hosted_registry_storage_access_modes`: ['ReadWriteOnce']
* `openshift_hosted_registry_storage_openstack_filesystem`: xfs
* `openshift_hosted_registry_storage_volume_size`: 10Gi

For a volume *you created*, you must also specify its **UUID** (it must be
the UUID, not the volume's name):

```
openshift_hosted_registry_storage_openstack_volumeID: e0ba2d73-d2f9-4514-a3b2-a0ced507fa05
```

If you want the volume *created automatically*, set the desired name instead:

```
openshift_hosted_registry_storage_volume_name: registry
```

The volume will be formatted automaticaly and it will be mounted to one of the
infra nodes when the registry pod gets started.


## Deploying At Scale

By default, heat stack outputs are resolved.  This may cause
problems in large scale deployments.  Querying heat stack can take
a long time and eventually time out.  The following setting in
`inventory/group_vars/all.yml` is recommended to prevent the timeouts:

* `openshift_openstack_resolve_heat_outputs`: False
