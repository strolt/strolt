import { generatePath, Params } from "react-router";

const rPrefix = (prefix: string) => (router?: string) => prefix + (router || "");

const rLk = (prefix: string) => rPrefix("/lk" + prefix);

const instances = rLk("/instances");
const _instances = {
  "instances.main": instances(),
  "instances.old": instances("/old"),
  "instances.instanceId.serviceId.taskId.destinationId.snapshotList": instances(
    "/:instanceId/:serviceId/:taskId/:destinationId/snapshots",
  ),
  "instances.proxyId.instanceId.serviceId.taskId.destinationId.proxySnapshotList": instances(
    "/:proxyId/:instanceId/:serviceId/:taskId/:destinationId/snapshots",
  ),
  "instances.instanceId.serviceId.taskId.destinationId.prune": instances(
    "/:instanceId/:serviceId/:taskId/:destinationId/prune",
  ),
  "instances.proxyId.instanceId.serviceId.taskId.destinationId.prune": instances(
    "/:proxyId/:instanceId/:serviceId/:taskId/:destinationId/prune",
  ),
  "instances.instanceId.serviceId.taskId.destinationId.stats": instances(
    "/:instanceId/:serviceId/:taskId/:destinationId/stats",
  ),
  "instances.proxyId.instanceId.serviceId.taskId.destinationId.proxyStats": instances(
    "/:proxyId/:instanceId/:serviceId/:taskId/:destinationId/stats",
  ),
};

const auth = rPrefix("/auth");
const _auth = {
  "auth.login": auth("/login"),
};

export const routerConstants = {
  main: "/",
  ..._instances,
  ..._auth,
};

export type ConstantsRouteType = keyof typeof routerConstants;

export const toPath = (constant: ConstantsRouteType) => routerConstants[constant];
export const toNavigate = (constant: ConstantsRouteType, params: Params = {}) => {
  return generatePath(routerConstants[constant], params);
};
