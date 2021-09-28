import React, {useState} from 'react';
import {Form, Input, Modal, Select} from "antd/lib/index";
import {isEmpty} from "../../utils/utils";

const {TextArea} = Input;

const accountTypes = [
    {text: '密码', value: 'custom'},
    {text: '密钥', value: 'private-key'},
    {text: '云服务商密钥', value: 'access-key'},
];

const cloudProvidersTypes = [
    {text: '腾讯云', value: 'qcloud'},
    {text: '阿里云', value: 'aliyun'},
];

const CredentialModal = ({title, visible, handleOk, handleCancel, confirmLoading, model}) => {

    const [form] = Form.useForm();

    const formItemLayout = {
        labelCol: {span: 6},
        wrapperCol: {span: 14},
    };

    if (model === null || model === undefined) {
        model = {}
    }

    if (isEmpty(model.type)) {
        model.type = 'custom';
    }

    for (let key in model) {
        if (model.hasOwnProperty(key)) {
            if (model[key] === '-') {
                model[key] = '';
            }
        }
    }

    let [type, setType] = useState(model.type);

    const handleAccountTypeChange = v => {
        setType(v);
        model.type = v;
    }

    const handleCloudProviderTypeChange = v => {
        model.provider = v;
    }

    return (

        <Modal
            title={title}
            visible={visible}
            maskClosable={false}

            onOk={() => {
                form
                    .validateFields()
                    .then(values => {
                        form.resetFields();
                        handleOk(values);
                    })
                    .catch(info => {

                    });
            }}
            onCancel={handleCancel}
            confirmLoading={confirmLoading}
            okText='确定'
            cancelText='取消'
        >

            <Form form={form} {...formItemLayout} initialValues={model}>
                <Form.Item name='id' noStyle>
                    <Input hidden={true}/>
                </Form.Item>

                <Form.Item label="凭证名称" name='name' rules={[{required: true, message: '请输入凭证名称'}]}>
                    <Input placeholder="请输入凭证名称"/>
                </Form.Item>

                <Form.Item label="类型" name='type' rules={[{required: true, message: '请选择类型'}]}>
                    <Select onChange={handleAccountTypeChange}>
                        {accountTypes.map(item => {
                            return (<Select.Option key={item.value} value={item.value}>{item.text}</Select.Option>)
                        })}
                    </Select>
                </Form.Item>

                {
                    type === 'private-key' ?
                        <>
                            <Form.Item label="授权账户" name='username'>
                                <Input placeholder="输入授权账户"/>
                            </Form.Item>

                            <Form.Item label="私钥" name='privateKey' rules={[{required: true, message: '请输入私钥'}]}>
                                <TextArea rows={4}/>
                            </Form.Item>
                            <Form.Item label="私钥密码" name='passphrase'>
                                <TextArea rows={1}/>
                            </Form.Item>
                        </>
                        :
                        ( type === "access-key" ? (
                            <>
                            {/* <Form.Item label="授权账户" name='username'>
                                <Input placeholder="输入授权账户"/>
                            </Form.Item> */}
                            <Form.Item label="云服务商" name='provider' rules={[{required: true, message: '请选择云服务商'}]}>
                                <Select onChange={handleCloudProviderTypeChange}>
                                    {cloudProvidersTypes.map(item => {
                                        return (<Select.Option key={item.value} value={item.value}>{item.text}</Select.Option>)
                                    })}
                                </Select>
                            </Form.Item>
                            <Form.Item label="ID" name='access_id'>
                                <Input placeholder="输入AccessKey ID"/>
                            </Form.Item>
                            <Form.Item label="Secret" name='access_secret'>
                                <Input.Password placeholder="输入AccessKey Secret"/>
                            </Form.Item>
                        </>

                        ) : (
                            <>
                            <Form.Item label="授权账户" name='username'>
                                <Input placeholder="输入授权账户"/>
                            </Form.Item>

                            <Form.Item label="授权密码" name='password'>
                                <Input.Password placeholder="输入授权密码"/>
                            </Form.Item>
                        </>

                            ) )
                        
                }

            </Form>
        </Modal>
    )
};

export default CredentialModal;
