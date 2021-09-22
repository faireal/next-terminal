import React, {Component} from 'react';
import {Card, Input, List, Spin} from "antd";
import Console from "../access/Console";

import './Command.css'
import request from "../../common/request";
import {message} from "antd/es";

const {Search} = Input;

class BatchCommand extends Component {

    commandRef = React.createRef();

    state = {
        webSockets: [],
        assets: [],
        active: undefined,
        loading: true
    }

    componentDidMount() {
        let params = new URLSearchParams(this.props.location.search);
        let assets = JSON.parse(params.get('assets'));
        let commandId = params.get('commandId');

        this.init(commandId, assets)
    }

    init = async (commandId, assets) => {

        let result = await request.get(`/apis/commands/${commandId}`);
        if (result['code'] !== 200) {
            message.error(result['message'], 10);
            this.setState({
                loading: false
            })
            return;
        }

        let command = result['data']['content'];
        this.setState({
            loading: false,
            command: command,
            assets: assets
        })
    }

    onPaneChange = activeKey => {
        this.setState({activeKey});
    };

    appendWebsocket = (webSocket) => {
        this.state.webSockets.push(webSocket);
    }

    render() {
        return (
            <>
                <Spin spinning={this.state.loading} tip='正在获取指令内容...'>
                    <div className="page-search">
                        <Search ref={this.commandRef} placeholder="请输入指令" onSearch={value => {
                            for (let i = 0; i < this.state.webSockets.length; i++) {
                                let ws = this.state.webSockets[i]['ws'];
                                if (ws.readyState === WebSocket.OPEN) {
                                    ws.send(JSON.stringify({
                                        type: 'data',
                                        content: value + String.fromCharCode(13)
                                    }));
                                }
                            }
                            this.commandRef.current.setValue('');
                        }} enterButton='执行'/>
                    </div>

                    <div className="page-card">
                        <List
                            grid={{gutter: 16, column: 2}}
                            dataSource={this.state.assets}
                            renderItem={item => (
                                <List.Item>
                                    <Card title={item.name}
                                          className={this.state.active === item['id'] ? 'command-active' : ''}
                                          onClick={() => {
                                              if (this.state.active === item['id']) {
                                                  this.setState({
                                                      active: undefined
                                                  })
                                              } else {
                                                  this.setState({
                                                      active: item['id']
                                                  })
                                              }
                                          }}
                                    >
                                        <Console assetId={item.id} command={this.state.command}
                                                 width={(window.innerWidth - 350) / 2}
                                                 height={420}
                                                 appendWebsocket={this.appendWebsocket}/>
                                    </Card>
                                </List.Item>
                            )}
                        />
                    </div>
                </Spin>

            </>
        );
    }
}

export default BatchCommand;