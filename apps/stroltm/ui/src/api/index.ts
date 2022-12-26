import axios from "axios";

import * as _api from "./generated/api";
import { Configuration } from "./generated/configuration";

export const axiosInstance = axios.create();

const basePath = window.location.origin;

export const config = new Configuration({ basePath });

const commonParams: any = [config, basePath, axiosInstance];

export const manager = new _api.ManagerApi(...commonParams);
export const auth = new _api.AuthApi(...commonParams);
export const global = new _api.GlobalApi(...commonParams);
