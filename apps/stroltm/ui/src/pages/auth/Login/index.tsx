import { useStores, observer } from "stores";

const Login = observer(() => {
  const { authStore } = useStores();

  return (
    <div>
      <button onClick={() => authStore.login("admin", "admin")}>Login</button>
    </div>
  );
});

export default Login;
