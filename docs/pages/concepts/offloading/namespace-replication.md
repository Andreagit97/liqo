---
title: Namespace replication
weight: 2
---

### The Liqo Namespace Model

The Liqo resource replication model is composed of several concurring logics responsible for different kind of objects to replicate.
One important example is represented by namespaces. 
In particular, namespace replication across multiple clusters implements a key mechanism to extend seamleassly a cluster to some others.

In the Liqo model, the namespace replication corresponds to a standard v1.Namespace and an associated [NamespaceOffloading](#) CR.
The NamespaceOffloading resource allows to select the clusters which are the clusters where the application contained inside the namespace can be scheduled.

In [a dedicated usage section](/usage/namespace_offloading), you can learn how to define and use a NamespaceOffloading resource activating the replication of the local namespace on peered clusters.

In this section, you may discover how Liqo namespace model works under the hood, undestanding which resources and controllers are activated when a new namespaceOffloading is manipulated.

#### Resources

Liqo namespace replication involves instances of two kinds of CRDs:

* First, the initial trigger is represented by the manipulation of a **NamespaceOffloading** (or the labelling of a namespace as explained in [the usage section](/usage/namespace_offloading)).
  On the one hand, the *NamespaceOffloading* spec describes the properties of the replication, such as the name of the replicated namespaces.
  On the other hand, the status collects the information about the actual status of remote namespaces (i.e. if the replication has succeeded).
* Second, **NamespaceMap** resources contain the list of namespaces associated to a specific node.
  The spec collects the list of desired namespaces for a specific remote cluster while the status updated information about their effective creation.
  Each *NamespaceMap* can be alimented by several *NamespaceOffloading* instances that are targeting the same cluster for a remote namespace.
  It is worth noting that the NamespaceMap status represent the source of truth to know the status of replicated namesapces on foreign clusters.

#### Controllers

The namespace replication logic is composed of several distinct controllers, responsible for different aspects of namespace controllers:

* **NamespaceOffloading Controller**: processes the NamespaceOffloading spec and outputs namespace entries to NamespaceMap specs.
* **NamespaceMap Controller**: enforces the namespace creation of the remote namespaces on remote clusters and updates the status of the NamespaceMaps.
* **OffloadingStatus Controller**: updates the status of the NamespaceOffloading resources by reading the NamespaceMaps statuses.

#### Workflow

In the next figure, you can observe a representation of the overall workflow steps of the replication process:

![](/images/namespace-replication/replication.png)

We can resume the steps in:

1. When the user creates/updates a **NamespaceOffloading** object in a Liqo-enabled namespace **(Step 1 in the figure)**, the Liqo logic processes the resource spec.

2. After having detected the virtual-nodes compliant with selector, a namespace creation request is entered in every NamespaceMap associated with a selected cluster **(Step 2 in the figure)**.
In particular, it considers the **ClusterSelector** field to select on which cluster the current namespace should be replicated.

More precisely, the controller responsible for this reconciliation is the **NamespaceOffloading Controller**. 
It processes the *NamespaceOffloading* spec fields, by inserting the namespace creation requests in the *spec.DesiredMapping* field of **NamespaceMap** instances of selected clusters.
The request format is an entry consisting of the local namespace name as a key and the name of the remote namespace as a value.
In addition, this operation has to be performed every time a new virtual-node joins the cluster after a new peering is established.

3. Once a *NamespaceMap* is updated, it is time to create the corresponding remote namespace. 
The *NamespaceMap Controller* reconciles the NamespaceMap resources, creating the remote namespaces and storing the operation results in the corresponding NamespaceMap.
More precisely, the occurred creation is saved in the *status.CurrentMapping* field, as you can see in **Step 3 of the figure**.
The result format is similar to request one: the key is the name of the local namespace, while the value is composed of the remote name and the actual remote namespace phase.

In addition to the actual creation of namespaces, the *NamespaceMap Controller* (1) periodically checks that each entry in the *spec.desiredMapping* field has an associated remote namespace and (2) performs health checks on the remote namespaces. The result of all its operations are stored in the *status.CurrentMapping* field in the NamespaceMap resource.

5. Liqo periodically checks that the requested remote namespaces are present.
Whenever it detects a change in the namespaces state, it immediately updates the NamespaceMap resources.
The NamespaceOffloading status is updated thanks to NamespaceMap status changes **(step 4 in the figure)**.
The **OffloadingStatus Controller** is responsible for this NamespaceOffloading status reconciliation. It periodically checks the status of all *NamespaceMaps* in the clusters, and for each NamespaceOffloading object, (1) it updates the RemoteNamespaceConditions with the actual remote namespaces status and (2)  it changes the global OffloadingPhase according to the previously set remote conditions. 

As already detailed in the [NamespaceOffloading status description](/usage/namespace_offloading/#check-the-namespaceoffloading-resource-status), the fields that provide the user with all the information about the replication phase are the RemoteNamespaceConditions and the OffloadingPhase **(step 5 in the figure)**. 

### Deletion workflow

When the user decides to delete the NamespaceOffloading resource, the *Offloading status controller* sets the OffloadingPhase of the NamespaceOffloading resource to Terminating. 
The corresponding entries are removed from the NamespaceMaps by the *NamespaceMap Controller*. 
Liqo reacts to this event by requesting the deletion of the remote namespaces that are no longer required.
In particular, the *NamespaceOffloading Controller* removes the creation requests from the *spec.desiredMapping* field of the NamespaceMap resources.

Consequently, the *NamespaceMap Controller* checks enforces the deletion of the remote namespaces that are no longer required.
When a remote namespace is deleted, the *NamespaceMap Controller* removes the corresponding entry from the *status.CurrentMapping* field of the NamespaceMap resource. 

When an entry is removed from the *status.CurrentMapping* field of one NamespaceMap resource, the *OffloadingStatus Controller* deletes the remote conditions associated with that namespace in the NamespaceOffloading resource.

Once all the remote namespaces have been removed and, therefore, all entries from the NamespaceMaps, then the NamespaceOffloading resource is finally removed, and the deletion process is complete. More precisely, the *OffloadingStatus Controller* sets the OffloadingPhase of the NamespaceOffloading resource to Terminating. The *NamespaceOffloading Controller* deletes the NamespaceOffloading when there are no more remote namespaces associated with this resource.