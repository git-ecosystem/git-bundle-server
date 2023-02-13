require 'asciidoctor'
require 'asciidoctor/extensions'

module GitBundleServer
  module Documentation
    class ManInlineMacro < Asciidoctor::Extensions::InlineMacroProcessor
      use_dsl

      named :man
      name_positional_attributes 'volnum'

      def process parent, target, attrs
        suffix = (volnum = attrs['volnum']) ? %((#{volnum})) : ''
        if parent.document.backend == 'manpage'
          # If we're building a manpage, bold the page name
          node = create_inline parent, :quoted, target, type: :strong
        else
          # Otherwise, leave the name as-provided
          node = create_inline parent, :quoted, target
        end
        create_inline parent, :quoted, %(#{node.convert}#{suffix})
      end
    end

    class UrlInlineMacro < Asciidoctor::Extensions::InlineMacroProcessor
      use_dsl

      named :url

      def process parent, target, attrs
        doc = parent.document
        if doc.backend == 'manpage'
          # If we're building a manpage, underline the name and escape the URL
          # to avoid autolinking (the .URL that Asciidoc creates doesn't
          # render correctly on all systems).
          escape = target.start_with?( 'http://', 'https://', 'ftp://', 'irc://', 'mailto://') ? '\\' : ''
          create_inline parent, :quoted, %(#{escape}#{target}), type: :emphasis
        else
          # Otherwise, pass through
          create_inline parent, :quoted, target
        end
      end
    end
  end
end

Asciidoctor::Extensions.register do
  inline_macro GitBundleServer::Documentation::ManInlineMacro
  inline_macro GitBundleServer::Documentation::UrlInlineMacro
end
