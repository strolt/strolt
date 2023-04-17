import { Suspense } from "react";

import { Outlet } from "react-router";

import { Link } from "components";

import { observer, useStores } from "stores";

import { Footer } from "./footer";
import { appConfigStore } from "stores/app-config.store";


export const Lk = () => {
  const { authStore } = useStores();
  return (
    <div>
      <button onClick={() => authStore.logout()}>Logout</button>

      <nav>
        <Link to="instances.main">Task List</Link>
        <Link to="instances.old">Instances</Link>
				<button onClick={()=>appConfigStore.toggleMode()}>toogle theme</button>
      </nav>
      <Suspense fallback="loading...">
        <Outlet />

        <Footer />
      </Suspense>
    </div>
  );
};
