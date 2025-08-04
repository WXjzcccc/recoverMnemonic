<template>
  <div class="container">
    <a-card title="以太坊钱包助记词顺序恢复工具" class="full-height-card">
      <a-form layout="vertical" class="form-container">
        <!-- 助记词输入框 -->
        <a-form-item label="助记词 (空格分隔的12个单词)">
          <a-input
              v-model:value="mnemonic"
              placeholder="请输入已知的助记词，单词之间用空格分隔"
              :disabled="isProcessing"
          />
        </a-form-item>

        <!-- 钱包地址输入框 -->
        <a-form-item label="钱包地址">
          <a-input
              v-model:value="walletAddress"
              placeholder="请输入钱包地址"
              :disabled="isProcessing"
          />
        </a-form-item>

        <!-- 按钮组 - 填满整行 -->
        <a-form-item class="button-group">
          <a-space :size="16" style="width: 100%">
            <a-button
                type="primary"
                @click="startProcessing"
                :disabled="isProcessing"
                style="flex: 1"
            >
              开始
            </a-button>
            <a-button
                danger
                @click="stopProcessing"
                :disabled="!isProcessing"
                style="flex: 1"
            >
              停止
            </a-button>
          </a-space>
        </a-form-item>

        <!-- 进度条 -->
        <a-form-item>
          <a-progress
              :percent="progressPercent"
              :stroke-color="progressColor"
              status="active"
              :stroke-width="20"
          />
        </a-form-item>

        <!-- 结果输出 - 支持内部滚动 -->
        <a-form-item label="恢复结果" class="result-container">
          <a-textarea
              v-model:value="result"
              placeholder="恢复结果将显示在这里"
              :rows="6"
              read-only
              class="result-textarea"
          />
        </a-form-item>
      </a-form>
    </a-card>
  </div>
</template>

<script setup>
import {ref, computed, onBeforeUnmount, watch, nextTick} from 'vue';
import { message } from 'ant-design-vue';
import {StopBruteForce,CrackMnemonic,GetProgress} from "../../wailsjs/go/main/App.js";

const mnemonic = ref('');
const walletAddress = ref('');
const isProcessing = ref(false);
const progressPercent = ref(0);
const progressInterval = ref(null);
const result = ref('');
const textareaRef = ref(null);

const progressColor = computed(() => {
  // 根据进度改变颜色
  if (progressPercent.value < 30) return '#ff4d4f';
  if (progressPercent.value < 70) return '#faad14';
  return '#52c41a';
});

const startProcessing = () => {
  if (mnemonic.value.split(" ").length !== 12) {
    message.warn("必须是12个助记词");
    return;
  }

  isProcessing.value = true;
  result.value = '开始验证...\n';
  progressPercent.value = 0;
  CrackMnemonic(mnemonic.value,walletAddress.value).then(cracked=>{
    clearInterval(progressInterval.value)
    result.value += cracked
    isProcessing.value = false
    progressPercent.value = 100
  })
  completeProcessing();
};

const stopProcessing = () => {
  StopBruteForce()
  isProcessing.value = false
};

const completeProcessing = () => {
  progressInterval.value = setInterval(()=>{
    GetProgress().then(pro=>{
      result.value += pro
      let pattern = /已处理: (\d+)\/(\d+),/
      let matches = pro.match(pattern)
      if(matches){
        var current = matches[1]
        var total = matches[2]
        progressPercent.value = Math.ceil(current/total*100)
      }
    })
  },1000)

};

onBeforeUnmount(() => {
  // 组件销毁时清除定时器
  if (progressInterval.value) {
    clearInterval(progressInterval.value);
  }
});

watch(result, () => {
  nextTick(() => {
    const textarea = document.querySelector('.result-textarea');
    if (textarea) {
      textarea.scrollTop = textarea.scrollHeight;
    }
  });
});
</script>

<style scoped>
.container {
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.full-height-card {
  flex: 1;
  display: flex;
  flex-direction: column;
  margin: 0;
  border-radius: 0;
  border: none;
}

.form-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 24px;
}

.button-group {
  margin-bottom: 16px;
}

.button-group .ant-space {
  display: flex;
}

.result-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0; /* 关键属性，允许内部元素收缩 */
}

.result-textarea {
  flex: 1;
  overflow-y: auto;
  resize: none;
}

/* 调整卡片内部布局 */
:deep(.ant-card-body) {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 0;
}

/* 调整表单间距 */
:deep(.ant-form-item) {
  margin-bottom: 16px;
}

:deep(.ant-form-item:last-child) {
  margin-bottom: 0;
}
</style>