import { makeStyles } from '@material-ui/styles';

export const useStyles = makeStyles({
  root: {
    border: 0,
    borderRadius: 3,
    height: 48,
    padding: '0 30px',
  },
  headerBg: {
    backgroundImage: 'linear-gradient(#FFFFFF, #E8E8E8)',
    fontSize: 2.5,
    fontWeight: 'normal',
    letterSpacing: '-1px',
    padding: '0.6em 0',
    color: "#9E9E9E",
    margin: 0
  },
  floatRight: {
    float: 'right',
    color: "#9E9E9E",
  },
  modalStyle: {
    position: 'absolute',
    top: '50%',
    left: '50%',
    transform: 'translate(-50%, -50%)',
    width: 400,
    bgcolor: 'background.paper',
    border: '2px solid #000',
    boxShadow: 24,
    p: 4,
  }
});
