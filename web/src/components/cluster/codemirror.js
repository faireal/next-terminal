import React from "react";
import { Controlled as CodeMirror } from "react-codemirror2";
import 'codemirror/lib/codemirror.css';
import 'codemirror/lib/codemirror.js';
import 'codemirror/mode/yaml/yaml';
import 'codemirror/theme/ambiance.css';
import 'codemirror/addon/selection/active-line';

class CodeMirrorWrapper extends React.Component {
  handleChange = (editor, data, value) => {
    this.props.onChange(value);
  };

  render() {
    const { value, onChange, ...restProps } = this.props;
    return (
      <CodeMirror
        value={value}
        options={{
            lineNumbers: true,
            theme: 'ambiance',
            mode: {
                name: 'text/x-yaml'
            },
        }}
        onBeforeChange={this.handleChange}
        {...restProps}
      />
    );
  }
}

export default CodeMirrorWrapper;