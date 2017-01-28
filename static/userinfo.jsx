/*
Display user info JSON from API endpoint
*/

class UserInfo extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      userInfo: {}
    };
  }

  componentDidMount() {
    axios.get('/auth/info')
      .then(res => {
        this.setState({ userInfo: res.data });
        console.log(this.state.userInfo)
      });
  }

  render() {
    return (
      <div>
        <pre>{JSON.stringify(this.state.userInfo, null, 2)}</pre>
      </div>
    );
  }
}

ReactDOM.render(<UserInfo />, document.getElementById('userinfo'));