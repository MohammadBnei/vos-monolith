import type { Meta, StoryObj } from '@storybook/vue3'
import DictionaryEntry from '@/components/word/DictionaryEntry.vue'

const meta: Meta<typeof DictionaryEntry> = {
  title: 'Components/Word/Dictionary Entry',
  component: DictionaryEntry,
  argTypes: {},
}

export default meta

type Story = StoryObj<typeof DictionaryEntry>

export const Mirador: Story = {
  args: {
    word: {
      id: '03f13341-e032-447e-bb82-2a81d38d4d39',
      text: 'mirador',
      language: 'fr',
      definitions: [
        {
          text: 'Tour de surveillance, poste d\'observation, surtout dans un camp de prisonniers ou une prison.',
          word_type: 'nom',
          examples: [
            'J\'avais essayé, grâce à l\'amitié d\'un sergent téléphoniste du mirador, de lui faire parvenir à mon tour un message rassurant… —(Romain Gary, La promesse de l\'aube, Folio)',
            'Le long de cours d\'eau qui creuse la région sauvage et montagneuse de la Transcarpatie, les gardes-frontières ne lésinent pas sur les moyens pour arrêter les fuyards: drones de surveillance thermiques, miradors, sentinelles postées en permanence aux affluents... —(Dans l\'ouest de l\'Ukraine, des contrebandiers devenus passeurs pour déserteurs, AFP - Boursorama, 16 aout 2024)',
          ],
          gender: 'masculin',
          pronunciation: '\\mi.ʁa.dɔʁ\\',
          language_specifics: {
            plural: 'miradors',
          },
        },
      ],
      etymology: '(1787)De l’espagnol mirador.',
      translations: {
        de: 'Wachtturm(de) masculin',
        en: 'watchtower(en)',
        es: 'mirador(es)',
      },
      search_terms: [
        'mirador',
        'miradors',
      ],
      lemma: 'mirador',
      created_at: '2025-03-15T23:26:34.202724Z',
      updated_at: '2025-03-15T23:26:34.595803Z',
    },
  },
}

export const Praxis: Story = {
  args: {
    word: {
      id: 'f5e6c772-ff81-4c5c-8031-0097e46c8e63',
      text: 'praxis',
      language: 'fr',
      definitions: [
        {
          text: 'Activité codifiée, manière générique de penser la transformation de l\'environnement.',
          word_type: 'nom',
          gender: 'féminin',
          pronunciation: '\\pʁak.sis\\',
        },
        {
          text: '(Sociologie) Ensemble des activités matérielles et intellectuelles des hommes qui contribuent à la transformation de la réalité sociale.',
          word_type: 'nom',
          gender: 'féminin',
          pronunciation: '\\pʁak.sis\\',
        },
        {
          text: 'Pratique.',
          word_type: 'nom',
          examples: [
            'Un ami qui [...] m\'a renvoyé à mes vraies forces. Ces forces qui logent davantage dans le monde intellectuel que dans la praxis. — (Odette Mainville, Le grand cahier de Jérôme, Fides, Montréal, 2020, p. 181)',
          ],
          gender: 'féminin',
          pronunciation: '\\pʁak.sis\\',
        },
        {
          text: '(Philosophie) Action dont l\'objet est le sujet lui-même.',
          word_type: 'nom',
          examples: [
            '[...] la finalité, toute mystique, de la praxis monastique consiste à faire dès maintenant l\'expérience de la vie céleste. — (Fabrizio Vecoli, « Le corps des moines », Argument, XXVII, 1, automne-hiver 2024-2025, page 9)',
          ],
          gender: 'féminin',
          pronunciation: '\\pʁak.sis\\',
        },
        {
          text: '(Psychologie) Activité psychique tournée vers un but, pratiquée pendant une psychanalyse.',
          word_type: 'nom',
          gender: 'féminin',
          pronunciation: '\\pʁak.sis\\',
        },
      ],
      etymology: '(Date à préciser) Du grec ancien πρᾶξις, praxis (« action »).',
      translations: {
        de: 'Praxis (de)',
        en: 'praxis (en)',
        pt: 'práxis (pt)',
      },
      search_terms: [
        'praxis',
      ],
      created_at: '2025-03-16T20:16:47.964654Z',
      updated_at: '2025-03-16T20:16:48.122566Z',
    },
  },
}

export const EmptyState: Story = {
  args: {
    word: undefined,
  },
}
