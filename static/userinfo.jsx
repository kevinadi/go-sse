/*
Display user info JSON from API endpoint
*/

class UserInfo extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      userInfo: null
    };
  }

  componentDidMount() {
    axios.get('/auth/info')
      .then(res => {
        this.setState({ userInfo: res.data });
      })
      .catch(err =>{
        if (err.response.status === 401) {
          this.setState({userInfo: null});
        }
      });
  }

  loginButton() {
    var loginButton;
    if (this.state.userInfo !== null) {
      return (<p><a href="/auth/logout">Logout</a></p>);
    } else {
      return (<p><a href="/auth/login">Login</a></p>);
    }
  }

  userInfoJSON() {
    if (this.state.userInfo !== null) {
      return (<pre>{JSON.stringify(this.state.userInfo, null, 2)}</pre>);
    } else {
      return (<pre>Not logged in</pre>);
    }
  }

  render() {
    return (
      <div>
        {this.loginButton()}
        {this.userInfoJSON()}
      </div>
    );
  }
}

ReactDOM.render(<UserInfo />, document.getElementById('userinfo'));