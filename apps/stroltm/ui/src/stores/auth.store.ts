import type { AxiosResponse } from "axios";
import type { IPromiseBasedObservable } from "mobx-utils";

import { makeAutoObservable, runInAction } from "mobx";
import { fromPromise } from "mobx-utils";

import type * as apiGenerated from "../api/generated";

import * as api from "../api";
import * as env from "../env";

const setApiAuthorization = (username: string, password: string) => {
  api.axiosInstance.defaults.auth = {
    password,
    username,
  };
};

export class AuthStore {
  isAuthorized = false;

  password = "";

  requestValidateStatus: IPromiseBasedObservable<
    AxiosResponse<apiGenerated.ApiuResultSuccess, any>
  > | null = null;

  username = "";
  constructor() {
    makeAutoObservable(this);
  }

  async login(username: string, password: string) {
    this.requestValidateStatus = fromPromise(
      api.auth.validate({
        password,
        username,
      }),
    );
    await this.requestValidateStatus;

    this.username = username;
    this.password = password;

    runInAction(() => {
      this.isAuthorized = true;
    });
    setApiAuthorization(username, password);
  }

  logout() {
    this.username = "";
    this.password = "";
    this.isAuthorized = false;
  }
}

export const authStore = new AuthStore();

if (env.isDevelopment) {
  document.addEventListener("DOMContentLoaded", () => {
    authStore.login("admin", "admin");
  });
}
