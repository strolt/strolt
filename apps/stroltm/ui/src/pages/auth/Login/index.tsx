import { Button, Form, Input, message } from "antd";

import { observer, useStores } from "stores";

import * as s from "./styles.css";

const Login = observer(() => {
  const { authStore } = useStores();

  const onFinish = async ({ password, username }: { password: string; username: string }) => {
    try {
      await authStore.login(username, password);
    } catch (error: any) {
      message.error(error?.message);
    }
  };

  return (
    <div className={s.authLayout}>
      <Form
        autoComplete="off"
        labelCol={{ span: 8 }}
        name="basic"
        onFinish={onFinish}
        wrapperCol={{ span: 16 }}
      >
        <Form.Item
          label="Username"
          name="username"
          rules={[{ message: "Please input your username!", required: true }]}
        >
          <Input />
        </Form.Item>

        <Form.Item
          label="Password"
          name="password"
          rules={[{ message: "Please input your password!", required: true }]}
        >
          <Input.Password />
        </Form.Item>

        <Form.Item wrapperCol={{ offset: 8, span: 16 }}>
          <Button
            htmlType="submit"
            loading={authStore.requestValidateStatus?.state === "pending"}
            type="primary"
          >
            Submit
          </Button>
        </Form.Item>
      </Form>
    </div>
  );
});

export default Login;
