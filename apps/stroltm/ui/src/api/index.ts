import axios from "axios";

import { message } from "antd";

import * as _api from "./generated/api";
import { Configuration } from "./generated/configuration";

export const axiosInstance = axios.create();

axiosInstance.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    if (error?.response?.data?.error) {
      message.error(error?.response?.data?.error);
    } else {
      message.error(error.message);
    }
    return error;
  },
);

const basePath = window.location.origin;

export const config = new Configuration({ basePath });

const commonParams: any = [config, basePath, axiosInstance];

export const managerDirect = new _api.ManagerDirectApi(...commonParams);
export const managerProxy = new _api.ManagerProxyApi(...commonParams);
export const manager = new _api.ManagerApi(...commonParams);
export const auth = new _api.AuthApi(...commonParams);
export const global = new _api.GlobalApi(...commonParams);
