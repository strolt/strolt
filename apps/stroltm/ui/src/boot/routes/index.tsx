import type { ReactNode } from "react";
import type { RouteObject } from "react-router";
import { Navigate } from "react-router";

import * as layouts from "layouts";
import { Login } from "pages/auth";
import { InstanceList, Prune, SnapshotList, Stats, TaskList } from "pages/lk";

import type { ConstantsRouteType } from "./constants";

import { toNavigate, toPath } from "./constants";

const r = (c: ConstantsRouteType, element: ReactNode) => ({
  element,
  path: toPath(c),
});

const instances = [
  r("instances.main", <TaskList />),
  r("instances.old", <InstanceList />),
  r("instances.instanceId.serviceId.taskId.destinationId.snapshotList", <SnapshotList />),
  r(
    "instances.proxyId.instanceId.serviceId.taskId.destinationId.proxySnapshotList",
    <SnapshotList />,
  ),
  r("instances.proxyId.instanceId.serviceId.taskId.destinationId.prune", <Prune />),
  r("instances.instanceId.serviceId.taskId.destinationId.prune", <Prune />),
  r("instances.instanceId.serviceId.taskId.destinationId.stats", <Stats />),
  r("instances.proxyId.instanceId.serviceId.taskId.destinationId.proxyStats", <Stats />),
];

const auth = [r("auth.login", <Login />)];

const routesLayoutLk = {
  children: [...instances],
  element: <layouts.Lk />,
  path: toPath("main"),
};

const routesLayoutAuth = {
  children: [...auth],
  element: <layouts.Auth />,
  path: toPath("main"),
};

export const routes = (isAuthorized: boolean): RouteObject[] => {
  if (isAuthorized) {
    return [
      routesLayoutLk,
      {
        element: <Navigate replace to={toNavigate("instances.main")} />,
        index: true,
        path: toPath("main"),
      },
      {
        element: <Navigate replace to={toNavigate("instances.main")} />,
        path: "*",
      },
    ];
  }

  return [
    routesLayoutAuth,
    {
      element: <Navigate replace to={toNavigate("auth.login")} />,
      index: true,
      path: toPath("main"),
    },
    {
      element: <Navigate replace to={toNavigate("auth.login")} />,
      path: "*",
    },
  ];
};
