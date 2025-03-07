# main.py
from textual.app import App, ComposeResult
from textual.widgets import Header, Footer, Tree, Static, Input, Button
from textual.containers import Vertical, Horizontal
import json
import paho.mqtt.client as mqtt
import uuid
import time
import random
from rich.text import Text
from rich.syntax import Syntax
from rich.panel import Panel
from rich.syntax import Syntax
from rich.console import Group

class MQTTDebuggerApp(App):
    CSS_PATH = "mqtt_debugger.tcss"

    BINDINGS = [
        ("q", "quit", "Quit"),
    ]
    
    REPLACE_MAP = {
        "gateway_sn": "SN123",
        "device_sn": "SN456",
    }
    
    def compose(self) -> ComposeResult:
        yield Header()
        yield Horizontal(
            Tree("Topics", id="topic_tree"),
            Vertical(
                Static("消息内容", id="message_content"),  # 更改为中文
                Input(placeholder="输入消息内容 (JSON格式)", id="message_input"),  # 更改为中文
                Horizontal(
                    Button("发送", id="send_button"),  # 更改为中文
                    Button("生成随机消息", id="generate_random_button"),  # 更改为中文
                    classes="button_container"  # 添加类名用于CSS定位
                ),
                id="right_panel"  # 添加ID用于CSS定位
            ),
        )
        yield Footer()

    def on_mount(self) -> None:
        self.selected_topic = None  # Initialize the selected_topic attribute
        self.message = None
        self.load_topics()
        self.client = mqtt.Client()
        self.client.connect("47.245.40.222", 1883, 60)  # 修改为你的 MQTT 服务器地址
        topic_tree = self.query_one(Tree)
        topic_tree.focus()

    def load_topics(self) -> None:
        topics = [
            "thing/product/device_sn/osd",
            "thing/product/device_sn/state",
            "thing/product/gateway_sn/services",
            "thing/product/gateway_sn/services_reply",
            "thing/product/gateway_sn/events",
            "thing/product/gateway_sn/events_reply",
            "thing/product/gateway_sn/requests",
            "thing/product/gateway_sn/requests_reply",
            "sys/product/gateway_sn/status",
            "sys/product/gateway_sn/status_reply",
            "thing/product/gateway_sn/property/set",
            "thing/product/gateway_sn/property/set_reply",
            "thing/product/gateway_sn/drc/up",
            "thing/product/gateway_sn/drc/down",
        ]
        topic_tree = self.query_one(Tree)
        for topic in topics:
            # Set the data attribute to the topic string when adding the leaf
            topic_tree.root.add_leaf(topic, data=topic)

    def on_tree_node_selected(self, event: Tree.NodeSelected) -> None:
        self.selected_topic = event.node.data
        # 替换占位
        for placeholder, value in self.REPLACE_MAP.items():
            self.selected_topic = self.selected_topic.replace(placeholder, value)
        # 更新消息内容区域
        self.query_one("#message_content", Static).update(f"Selected Topic: {self.selected_topic}")

    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "send_button":
            self.send_message()
        elif event.button.id == "generate_random_button":
            self.generate_random_message()

    def send_message(self) -> None:
        if not self.selected_topic:
            self.query_one("#message_content", Static).update("请先选择一个主题")  # 更改为中文
            return
        message = self.query_one("#message_input", Input).value
        try:
            payload = json.loads(message)
            self.client.publish(self.selected_topic, json.dumps(payload))
        except json.JSONDecodeError:
            self.query_one("#message_content", Static).update("Invalid JSON")

    def generate_random_message(self) -> None:
        if not self.selected_topic:
            return

        # 根据选中的topic生成随机消息
        if "osd" in self.selected_topic:
            payload = {
                "tid": str(uuid.uuid4()),
                "bid": str(uuid.uuid4()),
                "timestamp": int(time.time() * 1000),
                "data": {
                    "job_number": 492,
                    "acc_time": 1859010,
                    "activation_time": 0,
                    "maintain_status": {
                        "maintain_status_array": [
                            {
                                "state": 0,
                                "last_maintain_type": 17,
                                "last_maintain_time": 0,
                                "last_maintain_work_sorties": 0
                            }
                        ]
                    },
                    "electric_supply_voltage": 231,
                    "working_voltage": 25440,
                    "working_current": 1120,
                    "backup_battery": {
                        "voltage": 26631,
                        "temperature": 27.9,
                        "switch": 1
                    },
                    "drone_battery_maintenance_info": {
                        "maintenance_state": 0,
                        "maintenance_time_left": 0
                    }
                },
                "gateway": "SN123"
            } 
        elif "state" in self.selected_topic:
            payload = {
                "tid": str(uuid.uuid4()),
                "bid": str(uuid.uuid4()),
                "timestamp": int(time.time() * 1000),
                "gateway": "sn",
                "data": {}
            }
        elif "services" in self.selected_topic:
            payload = {
                "tid": str(uuid.uuid4()),
                "bid": str(uuid.uuid4()),
                "timestamp": int(time.time() * 1000),
                "gateway": "sn",
                "method": "some_method",
                "data": {}
            }
        elif "services_reply" in self.selected_topic:
            payload = {
                "tid": str(uuid.uuid4()),
                "bid": str(uuid.uuid4()),
                "timestamp": int(time.time() * 1000),
                "gateway": "sn",
                "method": "some_method",
                "data": {
                    "result": 0,
                    "output": {}
                }
            }
        elif "events" in self.selected_topic:
            payload = {
                "tid": str(uuid.uuid4()),
                "bid": str(uuid.uuid4()),
                "timestamp": int(time.time() * 1000),
                "need_reply": 0,
                "gateway": "sn",
                "method": "some_method",
                "data": {}
            }
        elif "events_reply" in self.selected_topic:
            payload = {
                "tid": str(uuid.uuid4()),
                "bid": str(uuid.uuid4()),
                "timestamp": int(time.time() * 1000),
                "gateway": "sn",
                "method": "some_method",
                "data": {
                    "result": 0
                }
            }
        elif "requests" in self.selected_topic:
            payload = {
                "tid": str(uuid.uuid4()),
                "bid": str(uuid.uuid4()),
                "timestamp": int(time.time() * 1000),
                "gateway": "sn",
                "method": "some_method",
                "data": {}
            }
        elif "requests_reply" in self.selected_topic:
            payload = {
                "tid": str(uuid.uuid4()),
                "bid": str(uuid.uuid4()),
                "timestamp": int(time.time() * 1000),
                "gateway": "sn",
                "method": "some_method",
                "data": {
                    "result": 0,
                    "output": {}
                }
            }
        elif "status" in self.selected_topic:
            payload = {
                "tid": str(uuid.uuid4()),
                "bid": str(uuid.uuid4()),
"method": "update_topo",
                "timestamp": int(time.time() * 1000),
                "data": {
                    "type": 98,
                    "sub_type": 0,
                    "device_secret": "secret",
                    "nonce": "nonce",
                    "version": 1,
                    "sub_devices": [
                        {
                            "sn": "SN456",
                            "type": 99,
                            "sub_type": 0,
                            "index": "A",
                            "device_secret": "secret",
                            "nonce": "nonce",
                            "version": 1
                        }
                    ]
                }
            }
        elif "property/set" in self.selected_topic:
            payload = {
                "tid": str(uuid.uuid4()),
                "bid": str(uuid.uuid4()),
                "timestamp": int(time.time() * 1000),
                "data": {
                    "some_property": "some_value"
                }
            }
        elif "property/set_reply" in self.selected_topic:
            payload = {
                "tid": str(uuid.uuid4()),
                "bid": str(uuid.uuid4()),
                "timestamp": int(time.time() * 1000),
                "data": {
                    "some_property": {
                        "result": 0
                    }
                }
            }
        elif "drc/up" in self.selected_topic:
            payload = {
                "method": "drone_control",
                "data": {
                    "result": 0,
                    "output": {
                        "seq": 1
                    }
                }
            }
        elif "drc/down" in self.selected_topic:
            payload = {
                "method": "drone_control",
                "data": {
                    "seq": 1,
                    "x": 2.34,
                    "y": -2.45,
                    "h": 2.76,
                    "w": 2.86
                }
            }
        else:
            payload = {}

        self.message = payload
        # 更新消息输入框
        self.query_one("#message_input", Input).value = json.dumps(payload, indent=4)
        
        # 创建带有主题信息的标题
        title = f"选中的主题: {self.selected_topic}"
        
        # 创建带有语法高亮的JSON
        json_syntax = Syntax(
            json.dumps(payload, indent=4),
            "json",
            theme="monokai",
            word_wrap=True
        )
        # 更新内容区域
        content = Panel(
            Group(json_syntax),
            title=title,
            border_style="blue"
        )
        
        # 更新内容区域
        self.query_one("#message_content", Static).update(content)

if __name__ == "__main__":
    app = MQTTDebuggerApp()
    app.run()