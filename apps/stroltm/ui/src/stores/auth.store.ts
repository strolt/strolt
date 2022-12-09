import { AxiosResponse } from "axios";
import { makeAutoObservable, runInAction } from "mobx";

import { fromPromise, IPromiseBasedObservable } from "mobx-utils";

import * as api from "../api";
import * as apiGenerated from "../api/generated";
import * as env from "../env";

const setApiAuthorization = (username: string, password: string) => {
  api.axiosInstance.defaults.auth = {
    username,
    password,
  };
};

export class AuthStore {
  constructor() {
    makeAutoObservable(this);
  }

  requestValidateStatus: IPromiseBasedObservable<
    AxiosResponse<apiGenerated.ApiuResultSuccess, any>
  > | null = null;

  isAuthorized = false;

  username = "";
  password = "";

  async login(username: string, password: string) {
    this.requestValidateStatus = fromPromise(
      api.auth.validate({
        username,
        password,
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
