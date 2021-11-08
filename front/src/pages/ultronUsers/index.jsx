import { useEffect, useState } from 'react';
import { useStyles } from '../components/makestyle';
import { Alert, DialogContent, TextField, Dialog, DialogActions, Divider, DialogTitle, Button } from '@material-ui/core';
import { Icon } from 'antd';

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
						label="ramp_up_period"
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

export const UltronUsers = ({ open, handleClose, setOpen,setStop }) => {
	const [planList, setPlanLists] = useState([]);
	const [message, setMessage] = useState('');

	useEffect(() => {
		open ? setMessage('') : '';
		open ? addOption() : '';
	}, [open]);

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
		fetch(`/api/v1/plan`, {
			method: 'POST',
			body: JSON.stringify(data),
		})
			.then(response => response.json())
			.then(function(res) {
				if (res && res.result) {
          setOpen(false);
          setStop(false)
				} else setMessage(res.error_message);
			});
	}

	return (
		<Dialog open={open} onClose={handleClose} scroll="body" fullWidth={true} maxWidth="sm">
			{message ? <Alert severity="error">{message}</Alert> : ''}
			<DialogTitle>Start New Plan</DialogTitle>
			<OptionsStagesConfig keyValue={planList} handleChange={handleChangeOption} removeOption={removeOption} />
			<DialogActions>
        <Button onClick={handleClose}>取消</Button>
        <Button onClick={addOption}>New Stage</Button>
				<Button onClick={() => handleSubmmit(planList)}>执行</Button>
			</DialogActions>
		</Dialog>
	);
};
