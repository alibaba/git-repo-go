#!/bin/sh
#
# Copyright (c) 2023 Jiang Xin
#

test_description='Test on test-tool xml-encode'

. lib/test-lib.sh

test_expect_success 'xml-encode test' '
	cat <<-EOF |
		<html>
		  <head>test-framework for R&D teams</head>
		  <body>
		  	Hello, world!
			你好，世界！
		  </body>
		</html>
	EOF
	test-tool xml-encode >actual &&
	printf "\n" >>actual &&
	cat >expect <<-EOF &&
		&lt;html&gt;&#x0a;  &lt;head&gt;test-framework for R&amp;D teams&lt;/head&gt;&#x0a;  &lt;body&gt;&#x0a;  &#x09;Hello, world!&#x0a;你好，世界！&#x0a;  &lt;/body&gt;&#x0a;&lt;/html&gt;&#x0a;
	EOF
	test_cmp expect actual
'

test_done
