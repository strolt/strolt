import axios from "axios";
import * as _api from "./generated/api";
import { Configuration } from "./generated/configuration";

export const axiosInstance = axios.create();

const basePath = window.location.origin;

export const config = new Configuration({ basePath });

const commonParams: any = [config, basePath, axiosInstance];

export const services = new _api.ServicesApi(...commonParams)
