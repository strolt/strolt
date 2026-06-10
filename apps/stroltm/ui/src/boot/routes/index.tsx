import type { ReactNode } from "react";
import type { RouteObject } from "react-router";
import { Navigate } from "react-router";

import * as layouts from "layouts";
import * as pages from "pages";

import type { ConstantsRouteType } from "./constants";

import { toNavigate, toPath } from "./constants";

const r = (c: ConstantsRouteType, element: ReactNode) => ({
  element,
  path: toPath(c),
});

const instances = [
  r("instances.main", <pages.lk.TaskList />),
  r("instances.old", <pages.lk.InstanceList />),
  r("instances.instanceId.serviceId.taskId.destinationId.snapshotList", <pages.lk.SnapshotList />),
  r(
    "instances.proxyId.instanceId.serviceId.taskId.destinationId.proxySnapshotList",
    <pages.lk.SnapshotList />,
  ),
  r("instances.proxyId.instanceId.serviceId.taskId.destinationId.prune", <pages.lk.Prune />),
  r("instances.instanceId.serviceId.taskId.destinationId.prune", <pages.lk.Prune />),
  r("instances.instanceId.serviceId.taskId.destinationId.stats", <pages.lk.Stats />),
  r("instances.proxyId.instanceId.serviceId.taskId.destinationId.proxyStats", <pages.lk.Stats />),
];

const auth = [r("auth.login", <pages.auth.Login />)];

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
