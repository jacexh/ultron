import { useState } from 'react';
import { useStyles } from '../components/makestyle';
import { Modal, Box, DialogContent, TextField, Dialog, DialogActions, DialogContentText, DialogTitle, Button } from '@material-ui/core';
import { connect } from 'dva';
import styles from './index.less';


const style = {
  position: 'absolute',
  top: '50%',
  left: '50%',
  transform: 'translate(-50%, -50%)',
  width: 400,
  bgcolor: 'background.paper',
  border: '2px solid #000',
  boxShadow: 24,
  p: 4,
};




export const UltronUsers = ({ open, handleClose, handleOk, flag = true }) => {
  const [host, setHost] = useState('')
  const [users, setUsers] = useState('')
  const [spawn, setSpawn] = useState('')

  const handleChange = (event) => {
    switch (event.target.id) {
      case 'users': setUsers(event.target.value); break;
      case 'spawn': setSpawn(event.target.value); break;
      case 'host': setHost(event.target.value); break;
    }
  };

  const handleSubmit = () => {
    //获取user host spawn
    console.log(host, users, spawn)
    handleOk(host, users, spawn)
  }

  return (
    <Dialog open={open} onClose={handleClose}>
      <DialogTitle>Start New Load Test</DialogTitle>
      <DialogContent>
        {/* <DialogContentText>
          描述信息输入。。。
        </DialogContentText> */}
        <TextField
          autoFocus
          margin="dense"
          id="users"
          label="并发用户数(设置模拟用户数)"
          fullWidth
          onChange={handleChange}
          variant="standard"
        />
        <TextField
          margin="dense"
          id="spawn"
          label="每秒产生（启动）的虚拟用户数"
          fullWidth
          variant="standard"
          onChange={handleChange}
        />
        {flag ? <TextField
          margin="dense"
          id="host"
          label="Host"
          fullWidth
          variant="standard"
          onChange={handleChange}
        /> : ''}
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose}>取消</Button>
        <Button onClick={handleSubmit}>开始运行</Button>
      </DialogActions>
    </Dialog>
  )


}