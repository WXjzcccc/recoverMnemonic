import {createApp} from 'vue'
import App from './App.vue'
import Antd from 'ant-design-vue';
import 'ant-design-vue/dist/reset.css';
import './style.css';

const app = createApp(App);
// createApp(App).mount('#app')
app.use(Antd).mount('#app');
