import React, {Component} from 'react';
import {Button, Card, Checkbox, Form, Input, Modal, Typography} from "antd";
import '../Login.css'
import request from "../../common/request";
import {message} from "antd/es";
import {withRouter} from "react-router-dom";
import {LockOutlined, OneToOneOutlined, UserOutlined} from '@ant-design/icons';

const {Title} = Typography;

class Oauth2 extends Component {

    formRef = React.createRef();
    totpInputRef = React.createRef();

    state = {
        inLogin: false,
        height: window.innerHeight,
        width: window.innerWidth,
        loginAccount: undefined,
        totpModalVisible: false,
        confirmLoading: false,
    };

    componentDidMount() {
        const m = this.props.match.params.type
        const q = this.props.location.search
        console.log(q)
        if ( m !== "callback" ) {
            this.autologin()
        } else {
            this.autocall(q)
        }
        window.addEventListener('resize', () => {
            this.setState({
                height: window.innerHeight,
                width: window.innerWidth
            })
        });
    }

    autologin = async value => {
        let result = await request.get("/oauth2/login", value);
        if (result.code === 200 ) {
            window.location.href = result.data
            return
        }
        message.error('接口异常')
    }

    autocall = async value => {
        let result = await request.get('/oauth2/callback' + value);
        if (result.code !== 200) {
            message.error(result.message)
            window.location.href = "/login"
            return
        }

        // 跳转登录
        sessionStorage.removeItem('current');
        sessionStorage.removeItem('openKeys');
        localStorage.setItem('X-Auth-Token', result['data']);
        // this.props.history.push();
        window.location.href = "/"
    }

    render() {
        return (
            <div 
                 style={{width: this.state.width, height: this.state.height, backgroundColor: '#F0F2F5'}}>
                <Card className='login-card' title={null}>
                    <div style={{textAlign: "center", margin: '15px auto 30px auto', color: '#F0F2F5'}}>
                        <Title level={1}>Next OPS Oauth</Title>
                    </div>
                </Card>
            </div>

        );
    }
}

export default withRouter(Oauth2);
