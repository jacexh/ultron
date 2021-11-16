import { useStyles } from '../components/makestyle';
import { notification, Icon } from 'antd';
import {
	Alert,
	Toolbar,
	Box,
	AppBar,
	Divider,
	Button,
	CircularProgress,
	Backdrop,
	Dialog,
	DialogActions,
	TextField,
	DialogTitle,
	DialogContent,
} from '@material-ui/core';
import { Edit, Stop } from '@material-ui/icons';
import styles from './index.css';
import { useEffect, useState } from 'react';
import parsePrometheusTextFormat from 'parse-prometheus-text-format';
import { UltronImage } from '../components/image';

const optionType = {
	strageConfig: {
		name: '',
		requests: '',
		duration: '',
		users: '',
		rampUpPeriod: '',
		minWait: '',
		maxWait: '',
	},
};

export const HeaderStatus = ({ title, textObj, flag = 1, color = '#5E5E5E', openEditUser = 'null', fontSize = '20' }) => {
	return (
		<>
			<div style={{ paddingLeft: 20, paddingRight: 20 }}>
				<span style={{ fontSize: 14, fontWeight: 600, fontFamily: 'Arial, Helvetica, sans-serif', color: '#666666' }}>{title}</span>
				<br />
				<span style={{ fontSize: 22, fontWeight: 600, fontFamily: 'Arial, Helvetica, sans-serif', color: color }}> {textObj}</span>
				{title == 'PLAN' ? (
					<a style={{ fontSize: 17, fontWeight: 400, fontFamily: 'Arial, Helvetica, sans-serif', color: '#6495ED' }} onClick={() => openEditUser()}>
						<Edit fontSize="small" />
						Edit
					</a>
				) : (
					''
				)}
			</div>
			<Divider orientation="vertical" variant="middle" flexItem />
		</>
	);
};

const OptionsStagesConfig = ({ keyValue, handleChange, removeOption }) => (
	<DialogContent>
		{keyValue &&
			keyValue.map((option, index) => (
				<div key={`option-${index}`}>
					{index == 0 ? (
						<TextField
							autoFocus
							size="small"
							value={option.name}
							margin="dense"
							id={`name${index}`}
							label="Plan名称"
							fullWidth
							variant={index == 0 ? 'outlined' : 'standard'}
							onChange={e => handleChange(e.target.value, index, 'name')}
						/>
					) : (
						<Divider>
							<h4>
								stage{index}&nbsp;
								<a onClick={e => removeOption(e, index)} style={{ color: '#EE4000', fontSize: 16 }}>
									<Icon type="minus-circle" />
								</a>
							</h4>
						</Divider>
					)}
					<TextField
						margin="dense"
						size="small"
						id={`users${index}`}
						value={option.users}
						label="用户数"
						onChange={e => handleChange(e.target.value, index, 'users')}
						variant="standard"
					/>
					<TextField
						margin="dense"
						size="small"
						id={`requests${index}`}
						value={option.requests}
						label="请求总数"
						onChange={e => handleChange(e.target.value, index, 'requests')}
						variant="standard"
					/>
					<TextField
						margin="dense"
						size="small"
						id={`rampUpPeriod${index}`}
						value={option.rampUpPeriod}
						label="准备时长"
						variant="standard"
						onChange={e => handleChange(e.target.value, index, 'rampUpPeriod')}
					/>
					<TextField
						margin="dense"
						size="small"
						id={`duration${index}`}
						value={option.duration}
						label="持续时长"
						variant="standard"
						onChange={e => handleChange(e.target.value, index, 'duration')}
					/>
					<TextField
						margin="dense"
						size="small"
						id={`minWait${index}`}
						label="最小等待时间"
						variant="standard"
						value={option.minWait}
						onChange={e => handleChange(e.target.value, index, 'minWait')}
					/>
					<TextField
						margin="dense"
						size="small"
						id={`maxWait${index}`}
						value={option.maxWait}
						label="最大等待时间"
						variant="standard"
						onChange={e => handleChange(e.target.value, index, 'maxWait')}
					/>
				</div>
			))}
	</DialogContent>
);

export const UltronHeader = ({ getMetrics, tableData }) => {
	const [open, setOpen] = useState(true);
	const [backStopDrop, setBackStopDrop] = useState(false);
	const [planList, setPlanLists] = useState([]);
	const [message, setMessage] = useState('');
	const [backDrop, setBackDrop] = useState(false);

	useEffect(() => {
		const timerId = setInterval(() => {
			tableData && tableData.tpsTotal ? '' : getMetrics();
		}, 4000);
		return () => {
			// 组件销毁时，清除定时器
			clearInterval(timerId);
		};
	});

	useEffect(() => {
		getMetrics();
	}, []);

	useEffect(() => {
		tableData && tableData.tpsTotal ? setBackStopDrop(false) : setOpen(false);
	}, [tableData]);

	const handleClose = () => {
		setPlanLists([]);
		setOpen(false);
	};
	const openEditUser = () => {
		setPlanLists([]);
		setOpen(true);
		localStorage.removeItem('chartData');
		localStorage.removeItem('tpsline');
	};

	useEffect(() => {
		open ? setMessage('') : '';
		open ? addOption() : '';
	}, [open]);

	function stopPlan() {
		setBackStopDrop(true);
		fetch(`/api/v1/plan`, {
			method: 'DELETE',
		})
			.then(response => response.json())
			.then(function(res) {
				if (res && res.result) '';
				else
					notification.error({
						message: `请求错误`,
						description: res.error_message,
						placement: 'bottomLeft',
					});
			});
	}

	function addOption() {
		var newValue = [...planList];
		var keyValue = optionType['strageConfig'];
		newValue.push({ ...keyValue });
		setPlanLists(newValue);
	}

	function removeOption(e, index) {
		const filterReault = planList.filter((echo, _index_) => _index_ !== index);
		setPlanLists(filterReault);
	}

	function handleChangeOption(e, index, type) {
		let newValue = planList;
		newValue[index][type] = e;
		setPlanLists([...newValue]);
	}

	function handleSubmmit(planObj) {
		var data = {};
		var config = [];
		planObj && planObj.length > 0
			? planObj.map((item, index) => {
					var c = {};
					index == 0 ? (data['name'] = item.name) : '';
					item['requests'] ? (c['requests'] = parseInt(item['requests'])) : '';
					item['duration'] ? (c['duration'] = item['duration']) : '';
					item['users'] ? (c['concurrent_users'] = parseInt(item['users'])) : '';
					item['rampUpPeriod'] ? (c['ramp_up_period'] = parseInt(item['rampUpPeriod'])) : '';
					item['maxWait'] ? (c['min_wait'] = item['maxWait']) : '';
					item['maxWait'] ? (c['max_wait'] = item['maxWait']) : '';
					config.push(c);
			  })
			: '';
		data['stages'] = config;
		setBackDrop(true);
		fetch(`/api/v1/plan`, {
			method: 'POST',
			body: JSON.stringify(data),
		})
			.then(response => response.json())
			.then(function(res) {
				if (res && res.result) isOver();
				else setMessage(res.error_message);
			});
	}

	function isOver() {
		fetch(`/metrics`, {
			method: 'GET',
		})
			.then(response => response.text())
			.then(function(res) {
				const metrics = parsePrometheusTextFormat(res);
				var f = false;
				for (var i of metrics) {
					if (i.name == 'ultron_attacker_tps_current') {
						getMetrics(); //确保开始了再调用getMetrics
						f = true;
						setOpen(false);
						setBackDrop(false);
						break;
					}
				}
				f ? '' : setTimeout(isOver, 1000);
			});
	}

	return (
		<>
			<Dialog scroll="body" fullWidth={true} maxWidth="sm" open={open} onClose={handleClose}>
				{message ? <Alert severity="error">{message}</Alert> : ''}
				<DialogTitle>Start New Plan</DialogTitle>
				<OptionsStagesConfig keyValue={planList} handleChange={handleChangeOption} removeOption={removeOption} />
				<DialogActions>
					<Button onClick={() => handleClose()}>取消</Button>
					<Button onClick={() => addOption()}>New Stage</Button>
					<Button onClick={() => handleSubmmit(planList)}>执行</Button>
				</DialogActions>
				<Backdrop sx={{ color: '#fff', zIndex: theme => theme.zIndex.drawer + 1 }} open={backDrop}>
					启动中...
					<CircularProgress color="inherit" />
				</Backdrop>
			</Dialog>
			<h1 className={styles.title}>
				<div>
					<AppBar position="fixed" className={useStyles().headerBg}>
						<div>
							<span style={{ paddingLeft: 25 }}>
								<UltronImage width="65" />
							</span>
							<span style={{ fontSize: 24, paddingTop: 10, fontWeight: 700, paddingLeft: 7, fontFamily: 'fantasy', color: '#404040' }}>Ultron</span>
							<Toolbar className={useStyles().floatRight}>
								<HeaderStatus title="PLAN" openEditUser={openEditUser} />
								<HeaderStatus title="USERS" textObj={tableData && tableData.users ? tableData.users : 0} />
								<HeaderStatus title="FAILURES" textObj={tableData && tableData.failureRatio ? tableData.failureRatio + '%' : 0} />
								{tableData && tableData.tpsTotal ? <HeaderStatus title="Total TPS" textObj={tableData.tpsTotal} /> : ''}
								&nbsp;&nbsp;
								{tableData && tableData.tpsTotal ? (
									''
								) : (
									<Button
										variant="contained"
										size="large"
										color="error"
										startIcon={<Stop />}
										onClick={() => {
											stopPlan();
										}}
									>
										STOP
									</Button>
								)}
								&nbsp;&nbsp;&nbsp;
							</Toolbar>
						</div>
						<br />
					</AppBar>
				</div>
			</h1>
			<Backdrop sx={{ color: '#fff', zIndex: theme => theme.zIndex.drawer + 1 }} open={backStopDrop}>
				停止中...
				<CircularProgress color="inherit" />
			</Backdrop>
		</>
	);
};
