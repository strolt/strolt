import { lazy } from "react";

export const InstanceList = lazy(() => import("./InstanceList"));
export const SnapshotList = lazy(() => import("./SnapshotList"));
export const Prune = lazy(() => import("./Prune"));
export const Stats = lazy(() => import("./Stats"));
export const TaskList = lazy(() => import("./TaskList"));
