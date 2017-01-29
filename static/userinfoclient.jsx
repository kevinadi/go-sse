/*
Display user info JSON from API endpoint
*/

class UserInfoClient extends React.Component {
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

  userInfoJSON() {
    if (auth2.isSignedIn.get()) {
        var profile = auth2.currentUser.get().getBasicProfile();
        // console.log('ID: ' + profile.getId());
        // console.log('Full Name: ' + profile.getName());
        // console.log('Given Name: ' + profile.getGivenName());
        // console.log('Family Name: ' + profile.getFamilyName());
        // console.log('Image URL: ' + profile.getImageUrl());
        // console.log('Email: ' + profile.getEmail());
        return (
            <pre>
            Client
            ID: {profile.getId()}
            Name: {profile.getName()}
            Given Name: {profile.getGivenName()}
            Family Name: {profile.getFamilyName()}
            Image URL: {profile.getImageUrl()}
            Email: {profile.getEmail()}
            </pre>
        )
    } else {
        return (<pre>Client: Not logged in</pre>)
    }
  }

  render() {
    return (
      <div>
        {this.userInfoJSON()}
      </div>
    );
  }
}

ReactDOM.render(<UserInfoClient />, document.getElementById('userinfoclient'));