import axios from "axios";
import * as _api from "./generated/api";
import { Configuration } from "./generated/configuration";

export const axiosInstance = axios.create();

const basePath = window.location.origin;

export const config = new Configuration({ basePath });

const commonParams: any = [config, basePath, axiosInstance];

export const instances = new _api.InstancesApi(...commonParams);
export const manager = new _api.ManagerApi(...commonParams);
export const auth = new _api.AuthApi(...commonParams);
