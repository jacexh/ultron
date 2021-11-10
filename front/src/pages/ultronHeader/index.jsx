import { useStyles } from '../components/makestyle';
import { Alert, Toolbar, Box, AppBar, Divider, Button } from '@material-ui/core';
import { AccessAlarm, Stop } from '@material-ui/icons';
import { connect } from 'dva';
import styles from './index.css';
import { UltronUsers } from '../ultronUsers/index';
import { useEffect, useState } from 'react';

export const HeaderStatus = ({ title, textObj, flag = 1, color = '#5E5E5E', openEditUser = 'null' }) => {
	return (
		<>
			<div style={{ paddingLeft: 20, paddingRight: 20 }}>
				<span style={{ fontSize: 14, fontWeight: 600, fontFamily: 'initial', color: '#666666' }}>{title}</span>
				<br />
				<span style={{ fontSize: 17, fontWeight: 600, fontFamily: 'monospace', color: color }}> {textObj}</span>
				{title == 'PLAN' ? (
					<a style={{ fontSize: 17, fontWeight: 400, fontFamily: 'monospace', color: '#6495ED' }} onClick={() => openEditUser()}>
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

export const UltronHeader = ({ getMetrics, tableData }) => {
	const [open, setOpen] = useState(true);
	const [metrics, setMetrics] = useState();
	const [stop, setStop] = useState(false);

	useEffect(() => {
		const timerId = setInterval(() => {
			getMetrics();
		}, 5000);
		return () => {
			// 组件销毁时，清除定时器
			clearInterval(timerId);
		};
	});

	useEffect(() => {
		getMetrics();
	}, []);

	const handleClose = () => {
		setOpen(false);
	};
	const openEditUser = () => {
		setOpen(true);
	};

	function stopPlan() {
		fetch(`/api/v1/plan`, {
			method: 'DELETE',
		})
			.then(response => response.json())
			.then(function(res) {
				console.log(res);
				if (res && res.result) {
					setStop(true);
					<Alert severity="success">success</Alert>;
				}
			});
	}

	return (
		<>
			<UltronUsers open={open} handleClose={handleClose} setOpen={setOpen} setStop={setStop} />
			<h1 className={styles.title}>
				<div>
					<AppBar position="fixed" className={useStyles().headerBg}>
						<div>
							<img src="./spaceman.png" width="75" style={{ paddingLeft: 25 }}></img>
							<span style={{ fontSize: 20, fontWeight: 700, paddingLeft: 7, fontFamily: 'fantasy', color: '#404040' }}>Ultron</span>
							<Toolbar className={useStyles().floatRight}>
								<HeaderStatus title="PLAN" openEditUser={openEditUser} />
								<HeaderStatus title="USERS" textObj={tableData && tableData.length > 0 ? tableData[0].users : 0} />
								{/* <HeaderStatus title="REQUESTS" textObj={tableData && tableData.length > 0 ? tableData[0].requests : 0} /> */}
								<HeaderStatus title="Total TPS" textObj={tableData && tableData.length > 0 ? tableData[0].tpsTotal : 0} />
								<HeaderStatus title="FAILURES" textObj={tableData && tableData.length > 0 ? tableData[0].failures : 0} />
								&nbsp;&nbsp;
								{stop ? (
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
		</>
	);
};
