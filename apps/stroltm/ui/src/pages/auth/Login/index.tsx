import { Button, Form, Input, message } from "antd";

import { useStores, observer } from "stores";

import * as s from "./styles.css";

const Login = observer(() => {
  const { authStore } = useStores();

  const onFinish = async ({ username, password }: { username: string; password: string }) => {
    try {
      await authStore.login(username, password);
    } catch (err: any) {
      message.error(err?.message);
    }
  };

  return (
    <div className={s.authLayout}>
      <Form
        name="basic"
        labelCol={{ span: 8 }}
        wrapperCol={{ span: 16 }}
        onFinish={onFinish}
        autoComplete="off"
      >
        <Form.Item
          label="Username"
          name="username"
          rules={[{ required: true, message: "Please input your username!" }]}
        >
          <Input />
        </Form.Item>

        <Form.Item
          label="Password"
          name="password"
          rules={[{ required: true, message: "Please input your password!" }]}
        >
          <Input.Password />
        </Form.Item>

        <Form.Item wrapperCol={{ offset: 8, span: 16 }}>
          <Button
            type="primary"
            htmlType="submit"
            loading={authStore.requestValidateStatus?.state === "pending"}
          >
            Submit
          </Button>
        </Form.Item>
      </Form>
    </div>
  );
});

export default Login;
