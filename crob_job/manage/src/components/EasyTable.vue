<script setup lang="ts">
import { onMounted, onUnmounted, ref } from "vue";
import axios from 'axios';
import { ElMessageBox } from 'element-plus'
import 'element-plus/theme-chalk/src/message-box.scss'
import 'element-plus/theme-chalk/src/message.scss'

interface CronListResponseData {
  JobName:string
  Cron:string
  Limit:number
	LastRun:Date
	Interval:number
	Status:number
	Description:string

  isEditing:boolean
  StatusStr:string
}

const tableData = ref<CronListResponseData[]>([])

GetTableData()

function GetTableData() {
  axios.get("/jobs").then((response)=>{
    if (response.status == 200){
      var rst:CronListResponseData[] = response.data
      rst.forEach((e,i) => {
        switch(e.Status) {
          case 1:
            rst[i].StatusStr = "准备就绪";
            break
          case 2:
            rst[i].StatusStr = "正在运行"
            break;
          case 0:
            rst[i].StatusStr = "已暂停"
            break;
          case -1:
            rst[i].StatusStr = "作业失败"
            break;
          case -2:
            rst[i].StatusStr = "失败暂停"
            break;
        }
      });
      tableData.value = response.data
    }
  })
}

// 使用ref来跟踪计时器
const intervalId = ref(0);
 
// 设置定时器函数
const refreshPageData = () => {
  intervalId.value = setInterval(() => {
    var isEditing = false
    tableData.value.forEach(e => {
      if (e.isEditing) {
        isEditing = true
        return
      }
    });
    if (!isEditing) {
      GetTableData()
    }
  }, 10000); // 5000毫秒后刷新页面
};

// 在组件挂载后启动定时器
onMounted(() => {
  refreshPageData();
});

// 在组件卸载前清除定时器
onUnmounted(() => {
  if (intervalId.value) {
    clearInterval(intervalId.value);
  }
});

const handleEdit = (row : CronListResponseData) => {
  if (row.isEditing) {
    axios.post("/jobs/update",{
      JobName:row.JobName,
      Cron:row.Cron,
      Limit:row.Limit
    }).then((response)=>{
      if (response.status == 200){
        GetTableData()
      }
    })
  }
  row.isEditing = !row.isEditing
}

const handlePauseOrResume = (row: CronListResponseData) => {
  var message = "你确定要" + (row.Status < 1 ? "重启任务:" : "暂停任务:") + row.JobName + "?"
  ElMessageBox.confirm(message).then( () => {
    if (row.Status < 1){
      axios.post("/jobs/"+row.JobName+"/resume").then((response)=>{
        if (response.status == 200){
          GetTableData()
        }
      })
    } else {
      axios.post("/jobs/"+row.JobName+"/pause").then((response)=>{
        if (response.status == 200){
          GetTableData()
        }
      })
    }
  })
}

</script>

<template>
  <div style="height: 36px; margin: 10px;"><el-button type="primary" @click="GetTableData()">刷新数据</el-button></div>
  <el-table :data="tableData" border style="width: 100%" row-style="height: 32px; line-height:32px;">
    <el-table-column prop="JobName" label="任务名称" width="200" />
    <el-table-column label="执行周期" width="160" >
      <template #default="scope">
        <el-input v-model="scope.row.Cron" :readonly="!scope.row.isEditing" />
      </template>
    </el-table-column>
    <el-table-column label="Limit" width="160" >
      <template #default="scope">
        <el-input v-model="scope.row.Limit" :readonly="!scope.row.isEditing" />
      </template>
    </el-table-column>
    <el-table-column prop="Status" label="状态" width="100">
      <template #default="scope">
        <el-popover effect="light" trigger="hover" placement="top" width="auto">
          <template #default>
            <div>Status: {{ scope.row.StatusStr }}</div>
            <div>Description: {{ scope.row.Description }}</div>
          </template>
          <template #reference>
            <el-tag v-if="scope.row.Status < 0" style="color: red; background-color: white; border-color: red;">{{ scope.row.StatusStr }}</el-tag>
            <el-tag v-else-if="scope.row.Status > 0" style="color: green; background-color: white; border-color: green;">{{ scope.row.StatusStr }}</el-tag>
            <el-tag v-else>{{ scope.row.StatusStr }}</el-tag>
          </template>
        </el-popover>
      </template>
    </el-table-column>
    <el-table-column label="操作" width="200">
      <template #default="{ row }">
        <el-button size="small" @click="handleEdit(row)" >{{ row.isEditing ? "保存" : "编辑"}}</el-button>
        <el-button v-if="row.isEditing" size="small">取消</el-button>
        <el-button size="small" :type="row.Status <1 ?'primary':'danger'" @click="handlePauseOrResume(row)">{{ row.Status <1 ? "重启" :"暂停"}}</el-button>
      </template>
    </el-table-column>
    <el-table-column prop="LastRun" label="上次执行时间" width="220" />
    <el-table-column prop="Interval" label="上次耗时(ms)" width="120" align="center" />
    <el-table-column prop="Description" show-overflow-tooltip label="上次运行描述" />
  </el-table>
</template>

<style scoped lang="scss">
body {
  background-color: white;
}
</style>