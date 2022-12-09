import * as layouts from "layouts";
import * as pages from "pages";

import { ReactNode } from "react";

import { RouteObject } from "react-router";
import { Navigate } from "react-router-dom";

import { toPath, toNavigate, ConstantsRouteType } from "./constants";

const r = (c: ConstantsRouteType, element: ReactNode) => ({
  path: toPath(c),
  element,
});

const instances = [
  r("instances.main", <pages.lk.InstanceList />),
  r("instances.instanceId.serviceId.taskId.destinationId.snapshotList", <pages.lk.SnapshotList />),
  r("instances.instanceId.serviceId.taskId.destinationId.prune", <pages.lk.Prune />),
];

const auth = [r("auth.login", <pages.auth.Login />)];

const routesLayoutLk = {
  path: toPath("main"),
  element: <layouts.Lk />,
  children: [...instances],
};

const routesLayoutAuth = {
  path: toPath("main"),
  element: <layouts.Auth />,
  children: [...auth],
};

export const routes = (isAuthorized: boolean): RouteObject[] => {
  if (isAuthorized) {
    return [
      routesLayoutLk,
      {
        path: toPath("main"),
        index: true,
        element: <Navigate replace to={toNavigate("instances.main")} />,
      },
      {
        path: "*",
        element: <Navigate replace to={toNavigate("instances.main")} />,
      },
    ];
  }

  return [
    routesLayoutAuth,
    {
      path: toPath("main"),
      index: true,
      element: <Navigate replace to={toNavigate("auth.login")} />,
    },
    {
      path: "*",
      element: <Navigate replace to={toNavigate("auth.login")} />,
    },
  ];
};
