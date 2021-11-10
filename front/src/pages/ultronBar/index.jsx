import React, { useState, useEffect } from 'react';
import { Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Tabs, Typography, Tab, Box, Paper } from '@material-ui/core';
import { useStyles } from '../components/makestyle';
import { styled } from '@material-ui/core/styles';
import { LineChart } from '../ultronBar/highcharttest';
import { tableCellClasses } from '@material-ui/core/TableCell';
import { Spin } from 'antd';

const StyledTableCell = styled(TableCell)(({ theme }) => ({
	[`&.${tableCellClasses.head}`]: {
		backgroundColor: theme.palette.common.black,
		color: theme.palette.common.white,
	},
	[`&.${tableCellClasses.body}`]: {
		fontSize: 14,
	},
}));

const StyledTableRow = styled(TableRow)(({ theme }) => ({
	'&:nth-of-type(odd)': {
		backgroundColor: theme.palette.action.hover,
	},
	// hide last border
	'&:last-child td, &:last-child th': {
		border: 0,
	},
}));

function TabPanel(props) {
	const { children, value, index, ...other } = props;
	return (
		<div role="tabpanel" hidden={value !== index} id={`simple-tabpanel-${index}`} aria-labelledby={`simple-tab-${index}`} {...other}>
			{value === index && (
				<Box sx={{ p: 3 }}>
					<Typography>{children}</Typography>
				</Box>
			)}
		</div>
	);
}

function a11yProps(index) {
	return {
		id: `simple-tab-${index}`,
		'aria-controls': `simple-tabpanel-${index}`,
	};
}

export const UltronBar = ({ tableData, lineData }, props) => {
	const { dispatch } = props;
	const [value, setValue] = useState(0);

	const handleChange = (event, newValue) => {
		setValue(newValue);
		switch (newValue) {
			case 0:
				{
				}
				break;
			case 1:
				{
				}
				break;
		}
	};
	return (
		<>
			<br />
			<Box sx={{ width: '100%', paddingTop: 4 }}>
				<Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
					<Tabs value={value} onChange={handleChange} centered className={useStyles().root}>
						<Tab label="Statistics" {...a11yProps(0)}></Tab>
						{lineData && lineData.length > 0 ? <Tab label="Charts" {...a11yProps(1)} /> : ''}
					</Tabs>
				</Box>
				<TabPanel value={value} index={0}>
					{/* <Spin spinning={tableData && tableData.length > 0 ? false : true}> */}
					<TableContainer component={Paper}>
						<Table sx={{ minWidth: 650 }} aria-label="simple table">
							<TableHead>
								<TableRow>
									<StyledTableCell>ATTACKER</StyledTableCell>
									<StyledTableCell align="center">MIN&nbsp;</StyledTableCell>
									<StyledTableCell align="center">P50&nbsp;</StyledTableCell>
									<StyledTableCell align="center">P60&nbsp;</StyledTableCell>
									<StyledTableCell align="center">P70&nbsp;</StyledTableCell>
									<StyledTableCell align="center">P80&nbsp;</StyledTableCell>
									<StyledTableCell align="center">P90&nbsp;</StyledTableCell>
									<StyledTableCell align="center">P95&nbsp;</StyledTableCell>
									<StyledTableCell align="center">P97&nbsp;</StyledTableCell>
									<StyledTableCell align="center">P98&nbsp;</StyledTableCell>
									<StyledTableCell align="center">P99&nbsp;</StyledTableCell>
									<StyledTableCell align="center">MAX&nbsp;</StyledTableCell>
									<StyledTableCell align="center">AVG&nbsp;</StyledTableCell>
									<StyledTableCell align="center">REQUESTS&nbsp;</StyledTableCell>
									<StyledTableCell align="center">TPS&nbsp;</StyledTableCell>
								</TableRow>
							</TableHead>
							<TableBody>
								{tableData.map((row, index) => (
									<StyledTableRow key={index}>
										<StyledTableCell component="th" scope="row">
											{row.attacker}
										</StyledTableCell>
										<StyledTableCell align="center">{row.MIN}</StyledTableCell>
										<StyledTableCell align="center">{row.P50}</StyledTableCell>
										<StyledTableCell align="center">{row.P60}</StyledTableCell>
										<StyledTableCell align="center">{row.P70}</StyledTableCell>
										<StyledTableCell align="center">{row.P80}</StyledTableCell>
										<StyledTableCell align="center">{row.P90}</StyledTableCell>
										<StyledTableCell align="center">{row.P95}</StyledTableCell>
										<StyledTableCell align="center">{row.P97}</StyledTableCell>
										<StyledTableCell align="center">{row.P98}</StyledTableCell>
										<StyledTableCell align="center">{row.P99}</StyledTableCell>
										<StyledTableCell align="center">{row.MAX}</StyledTableCell>
										<StyledTableCell align="center">{row.AVG}</StyledTableCell>
										<StyledTableCell align="center">{row.requests}</StyledTableCell>
										<StyledTableCell align="center">{row.tpsCurrent}</StyledTableCell>
									</StyledTableRow>
								))}
							</TableBody>
						</Table>
					</TableContainer>
					{/* </Spin> */}
				</TabPanel>
				<TabPanel value={value} index={1}>
					<LineChart lineData={lineData} />
				</TabPanel>
				<TabPanel value={value} index={2}></TabPanel>
				<TabPanel value={value} index={3}>
					Item four
				</TabPanel>
				<TabPanel value={value} index={4}>
					Item five
				</TabPanel>
			</Box>
		</>
	);
};
